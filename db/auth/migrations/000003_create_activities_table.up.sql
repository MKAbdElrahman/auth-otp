CREATE TABLE activities (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(15) NOT NULL,
    type VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (phone_number) REFERENCES users(phone_number) ON DELETE CASCADE
);