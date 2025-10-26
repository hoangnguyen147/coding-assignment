I have created 2 Go Map to store transaction information and user balances. I use mutex to protect the shared data access. I have also created a HTTP handler to process payment requests.
In reality, we will never use Go Map to store transaction information and user balances. We will use database (Postgres, MySQL, Redis, etc.) to store them. But for the simplicity of the exercise, I have used Go Map. This approach won't work if we have multiple instances of the service. In that case, we will use Redis or other distributed cache to store the data.

The main idea of this project is using mutex to protect shared data access (pessimistic locking). I have also implemented idempotency to ensure that a transaction is processed only once even if the request is sent multiple times.

The amount can be positive or negative. If it's positive, it will be added from the user balance. If it's negative, it will be deducted to the user balance.
