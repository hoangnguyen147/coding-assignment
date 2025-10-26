-- Write a query to get the last 10 transactions of a given user (user_id = 123)
SELECT id, user_id, amount, created_at
FROM transactions
WHERE user_id = 123
ORDER BY created_at DESC
LIMIT 10;

-- Suggest one index that would improve query performance.
-- Index on user_id and created_at, the query will use the index to filter by user_id first and then sort by created_at
CREATE INDEX idx_user_created ON transactions(user_id, created_at);
-- I won't mention write overhead due to indexing because it's not the focus of the question.