CREATE TABLE public.countmetric (id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,name varchar(100),val bigint,created timestamp default now());
	                                  
CREATE TABLE public.gaugemetric (id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,name varchar(100), val numeric(100,32), created timestamp default now());
