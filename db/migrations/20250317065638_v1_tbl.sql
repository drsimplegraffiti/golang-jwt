-- migrate:up
-- Active: 1737950469124@@127.0.0.1@5432@limo
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,                   -- Use SERIAL instead of AUTO_INCREMENT
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- User profile table (associated with users)
CREATE TABLE IF NOT EXISTS user_profiles (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    profile_image VARCHAR(255), -- URL to the profile image (e.g., Cloudinary link)
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS blogs (
    id SERIAL PRIMARY KEY,                   -- Use SERIAL instead of AUTO_INCREMENT
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    user_id INT NOT NULL,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);


-- migrate:down

