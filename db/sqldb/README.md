# Placeholders

## 1. Simple Placeholders
For simple placeholders, we use `?` and then convert them to each DBMS with conversion function.

### MySQL
No Conversion Required

### PostgreSQL
`?` -> `$n` where n = 1, 2, 3, ...

### MS SQL
`?` -> `@n` where n = 1, 2, 3, ...

### Oracle
`?` -> `:n` where n = 1, 2, 3, ...

## 2. Dynamic Placeholders
A dynamic placeholder is a notation used to represent a variable number of placeholders. It uses a single character `@`, which can be converted—via a conversion function—into forms like `?, ?, ..., ?` or `$k, $k+1, ..., $n`, depending on the DBMS.

