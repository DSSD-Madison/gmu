-- Create the 'Region' table
CREATE TABLE Region (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- Create the 'documents' table
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    abstract TEXT,
    region_id INT REFERENCES Region(id) ON DELETE SET NULL,
    category VARCHAR(100),
    publish_date DATE,
    source VARCHAR(255),
    image_link VARCHAR(255),
    pdf_link VARCHAR(255),
    s3_link VARCHAR(255),
    last_modified TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create the 'authors' table
CREATE TABLE authors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- Create the 'writtenBy' table (Many-to-Many between 'documents' and 'authors')
CREATE TABLE writtenBy (
    article_id INT REFERENCES documents(id) ON DELETE CASCADE,
    author_id INT REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, author_id)
);

-- Create the 'keywords' table
CREATE TABLE keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL UNIQUE
);

-- Create the 'keywordReference' table (Many-to-Many between 'documents' and 'keywords')
CREATE TABLE keywordReference (
    article_id INT REFERENCES documents(id) ON DELETE CASCADE,
    keyword_id INT REFERENCES keywords(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, keyword_id)
);



CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    abstract TEXT,
    region INT REFERENCES Region(id) ON DELETE SET NULL,
    category VARCHAR(100),
    publish_date DATE,
    source VARCHAR(255),
    image_id VARCHAR(100),
    pdf_id VARCHAR(100),
    orig_link VARCHAR(500),
    last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

