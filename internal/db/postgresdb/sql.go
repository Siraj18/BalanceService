package postgresdb

const initSchema = `
				DROP TABLE IF EXISTS reserves;
				DROP TABLE IF EXISTS transactions;
				DROP TABLE IF EXISTS users;
				
				CREATE TABLE IF NOT EXISTS users 
				(
					id  	UUID PRIMARY KEY,
					balance DECIMAL(10, 2) DEFAULT 0 CHECK (balance >= 0)
				);
				CREATE TABLE IF NOT EXISTS transactions
				(
					id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					to_id      UUID REFERENCES users(id),
					from_id    UUID REFERENCES users(id),
					money      DECIMAL(10, 2) NOT NULL,
					operation  TEXT NOT NULL,
					created_at TIMESTAMP DEFAULT now()
				);
				CREATE TABLE IF NOT EXISTS reserves
				(
					id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id       UUID REFERENCES users(id),
					service_id    TEXT NOT NULL,
					order_id      TEXT NOT NULL,
					amount        DECIMAL(10, 2) NOT NULL,
					status        TEXT NOT NULL,
					created_at    TIMESTAMP DEFAULT now(),
					recognized_at TIMESTAMP DEFAULT NULL
				);
				CREATE INDEX ON transactions (to_id);
				CREATE INDEX ON transactions (from_id);
`

const addUserSql = `
				INSERT INTO users VALUES ($1, 0);
`

const updateUserBalanceSql = `
				UPDATE users SET balance=balance + $2
				WHERE id=$1
				RETURNING *;
`

const getUserSql = `
				SELECT id, balance FROM users
				WHERE id=$1;
`

const addTransactionsSql = `
				INSERT INTO transactions (to_id, from_id, money, operation, created_at)
				VALUES ($1, $2, $3, $4, $5);
`

const getAllTransactionsSql = `
				SELECT id, to_id, from_id, money, operation, created_at FROM transactions
				WHERE to_id=$1 OR from_id=$1
				%s
				LIMIT $2
				OFFSET $3;
`

const addReserveSql = `
				INSERT INTO reserves (user_id, service_id, order_id, amount, status, created_at)
				VALUES ($1, $2, $3, $4, $5, $6);
`

const getReserveSql = `
				SELECT id, user_id, service_id, order_id, amount, status, created_at, recognized_at FROM reserves
				WHERE user_id=$1 and service_id=$2 and order_id=$3 and amount=$4;
`

const getReserveForReportSql = `
				SELECT id, user_id, service_id, order_id, amount, status, created_at, recognized_at FROM reserves
				WHERE status=$1 and date_part('year', recognized_at)=$2 and date_part('month', recognized_at)=$3;
`

const updateReserveStatus = `
				UPDATE reserves SET status=$2, recognized_at=$3
				WHERE id=$1;
`
