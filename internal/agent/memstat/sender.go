package memstat

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/fdanis/ygtrack/internal/helpers"
	"github.com/fdanis/ygtrack/internal/server/models"
	"golang.org/x/sync/errgroup"
)

type SenderMetric struct {
	send         func(client *http.Client, url string, header map[string]string, data *bytes.Buffer) error
	httpclient   *http.Client
	SendersCount int
	PublicKey    *rsa.PublicKey
}

func NewSenderMetric() *SenderMetric {
	s := &SenderMetric{}
	s.send = post
	s.httpclient = &http.Client{}
	s.SendersCount = 10
	return s
}

func (s *SenderMetric) Encription(data *bytes.Buffer) (*bytes.Buffer, error) {
	if s.PublicKey == nil {
		return data, nil
	}
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		s.PublicKey,
		data.Bytes(),
		nil)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(encryptedBytes), nil
}

func (s *SenderMetric) Send(url string, metrics []*models.Metrics) {
	g := errgroup.Group{}
	recordCh := make(chan *bytes.Buffer)
	for i := 0; i < s.SendersCount; i++ {
		w := &sendWorker{
			ch: recordCh,
			send: func(data *bytes.Buffer) error {
				data, err := s.Encription(data)
				if err != nil {
					return err
				}
				return s.send(s.httpclient, url, map[string]string{"Content-Type": "application/json"}, data)
			},
		}
		g.Go(w.do)
	}
	//здесь проще было бы наверно использовать сначала Marshal для всех метрик
	//а уже потом отправлять в несколько потоков запросы
	//но хотел попробывать именно этот патерн
	w := &marshalWorker{ch: recordCh, list: metrics}
	err := w.do()
	if err != nil {
		log.Println(err)
	}
	close(recordCh)
	err = g.Wait()
	if err != nil {
		log.Println(err)
	}
}

func (s *SenderMetric) SendBatch(url string, metrics models.Metrics) {
	d, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("could not do json.marshal %v", err)
		return
	}

	var buf *bytes.Buffer
	w := io.Writer(buf)
	gz := helpers.GetPool().GetWriter(w)
	defer helpers.GetPool().PutWriter(gz)
	_, err = gz.Write(d)
	if err != nil {
		log.Println(err)
	}
	gz.Flush()

	buf, err = s.Encription(buf)
	if err != nil {
		log.Println("can not encrypt")
	}

	err = s.send(s.httpclient, url, map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip"}, buf)
	if err != nil {
		log.Println("can not send batch")
	}
}

type marshalWorker struct {
	list []*models.Metrics
	ch   chan *bytes.Buffer
}

func (w *marshalWorker) do() error {
	for _, item := range w.list {
		d, err := json.Marshal(item)
		if err != nil {
			log.Printf("could marshal %v", err)
			return err
		}
		w.ch <- bytes.NewBuffer(d)
	}
	return nil
}

type sendWorker struct {
	ch   chan *bytes.Buffer
	send func(data *bytes.Buffer) error
}

func (w *sendWorker) do() error {
	for data := range w.ch {
		err := w.send(data)
		if err != nil {
			log.Print(err)
		}
	}
	return nil
}

func post(client *http.Client, url string, header map[string]string, data *bytes.Buffer) error {
	r, err := http.NewRequest("POST", url, data)
	if err != nil {
		return err
	}
	for k, v := range header {
		r.Header.Add(k, v)
	}
	res, err := client.Do(r)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got wrong http status (%d)", res.StatusCode)
	}
	return nil
}
