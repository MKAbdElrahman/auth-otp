CREATE TABLE otps (
    phone_number VARCHAR(15) PRIMARY KEY,
    code VARCHAR(10) NOT NULL,
    expiration TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (phone_number) REFERENCES users(phone_number) ON DELETE CASCADE
);