# TODO

## general
- [x] CRUD for accounts/transactions 
- [x] is expense support for transactions (rendering, input) - expenses are just transactions with negative amount lol, easy (actually not such a great idea, but it's for me and I am okay with it)
- [x] transactions search (`LIKE %<value>%` by account name, description, categories)
- [x] rework db logic to support both db handle and transaction handle somehow - used interface to conform both sql.Tx and sql.DB
- [x] implement db transactions on transaction create/update (to update account as well) 
- [ ] main screen with stats
- [ ] rework layout for mobile screen
- [ ] make new account card and edit account card the same thing, as I did for transactions
- [ ] make new account form behave the same way as for transactions (send from server on request)
- [ ] roll out custom transaction div-based card like:

```
[Category]  amount  very long
Account     date    description
```

- [x] create `exchange` category (might not affects the cash flow stats? will figure it out later)
- [ ] active search for accounts
- [ ] accounts filtering by currency
- [ ] support for terms (somehow) in active search

## transactions filtering
- [x] datetime ranges
- [ ] expense/income 
- [x] categories, accounts - not really, but handled by active search basically





