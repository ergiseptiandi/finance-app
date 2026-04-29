# Postman Collection

This folder contains the Postman files for all API features in this project.

## Files

- `finance-backend.postman_collection.json`: the full endpoint collection
- `finance-backend.local.postman_environment.json`: the local environment with base variables

The collection already includes placeholder filters for transactions, dashboard, and reports, including `month`, `year`, `start_date`, `end_date`, and `budget_amount`.
For requests with multiple filter modes, fill in only the mode you want to use.

Dashboard summary note: `total_balance` is the running balance, while `period_balance` is the balance for the selected period.

## How To Use

1. Import both files into Postman.
2. Select the `Finance Backend Local` environment.
3. Run the `Auth > Login` request.

## Auto Save Token

The `Auth > Login` request includes a test script that automatically stores:

- `access_token`
- `refresh_token`
- `token_type`
- `user_id`

The token is saved to both **collection variables** and **environment variables**, so protected requests can be used immediately.
