# order service for tests

make deps-docker-pg

execute sql script
```sql

create table public.orders
(
    id           bigserial
        constraint orders_pk
            primary key,
    client_id    bigint,
    number       integer,
    order_number varchar(255)
);

create unique index orders_order_number_uindex
    on public.orders (order_number);


create table public.orders_id
(
    client_id    bigint  not null
        constraint orders_id_pk
            primary key,
    orders_count integer not null
);

insert into orders_id (client_id, orders_count) values (1::bigint, 0::integer);

update orders set order_number = concat(client_id, '-', number) where order_number is null;

truncate orders;
alter sequence orders_id_seq restart with 1;
update orders_id set orders_count = 0 where 1=1;


select number from orders where client_id = 1 order by id desc limit 1;
```

go run ./cmd/orders/main.go

see this code appImpl.Create


```go
// work correct
// http_req_failed................: 0.00%  ✓ 0          ✗ 9556
err := a.repo.WithCte(ctx, request.ClientId)
```
```go
// work with errors
// http_req_failed................: 80.27% ✓ 6759       ✗ 1661
err := a.repo.WithSelect(ctx, request.ClientId)
```

k6 run tank/k6.js