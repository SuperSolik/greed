# TODO

## general
- [x] CRUD for accounts/transactions 
- [x] is expense support for transactions (rendering, input) - expenses are just transactions with negative amount lol, easy (actually not such a great idea, but it's for me and I am okay with it)
- [x] transactions search (`LIKE %<value>%` by account name, description, categories)
- [x] rework db logic to support both db handle and transaction handle somehow - used interface to conform both sql.Tx and sql.DB
- [x] implement db transactions on transaction create/update (to update account as well) 
- [x] main screen with stats
- [ ] rework layout for mobile screen
    - [ ] center content
    - [ ] roll out custom account div-based card
    - [ ] roll out custom transaction div-based card like:
    
    ```
    [Category]  amount  very long
    Account     date    description
    ```
- [ ] make new account card and edit account card the same thing, as I did for transactions
- [x] make new account form behave the same way as for transactions (send from server on request)

- [x] create `exchange` category (might not affects the cash flow stats? will figure it out later)
- [ ] active search for accounts
- [ ] accounts filtering by currency
- [ ] support for terms (somehow) in active search
- [ ] auth (signin, signup, sessions) + user based logic
- [ ] db indices on searchable fields
- [ ] create `<relative-time></relative-time>` web component to render local time (instead of hyperscript hack), inspiration - https://www.npmjs.com/package/@github/relative-time-element

## transactions filtering
- [x] datetime ranges
- [x] expense/income 
- [x] categories, accounts - not really, but handled by active search basically





