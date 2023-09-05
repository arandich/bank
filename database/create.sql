CREATE TABLE clients (
                         id SERIAL PRIMARY KEY,
                         name VARCHAR(255) NOT NULL,
                         token VARCHAR(255),
                         balance DECIMAL(10, 2) NOT NULL
);

CREATE TABLE transactions (
                              id varchar PRIMARY KEY,
                              sender_id INT NOT NULL,
                              receiver_id INT NOT NULL,
                              status VARCHAR(255) NOT NULL,
                              amount DECIMAL(10, 2) NOT NULL,
                              created_at TIMESTAMP DEFAULT NOW(),
                              FOREIGN KEY (sender_id) REFERENCES clients(id),
                              FOREIGN KEY (receiver_id) REFERENCES clients(id)
);

INSERT INTO clients (name, balance, token) VALUES ('John Doe', 1000.00, 'testToken');
INSERT INTO clients (name, balance,token) VALUES ('Jane Doe', 1000.00, 'testToken2');
