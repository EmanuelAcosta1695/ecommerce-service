CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT false
);

ALTER TABLE orders
    ADD COLUMN user_id INT NOT NULL,
    ADD CONSTRAINT fk_user
        FOREIGN KEY (user_id)
        REFERENCES users(id);