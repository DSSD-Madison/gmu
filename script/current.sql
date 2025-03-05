-- Table: documents (Main Storage Table)
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    abstract TEXT,
    region INT REFERENCES regions(id) ON DELETE SET NULL,
    category VARCHAR(100),
    publish_date DATE,
    source VARCHAR(255),
    image_id VARCHAR(100),
    pdf_id VARCHAR(100),
    orig_link VARCHAR(500),
    last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: regions (Geographic Data)
CREATE TABLE regions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- Table: authors
CREATE TABLE authors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

-- Table: writtenby (Many-to-Many Relationship between documents and authors)
CREATE TABLE writtenby (
    article_id INT REFERENCES documents(id) ON DELETE CASCADE,
    author_id INT REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, author_id)
);

-- Table: keywords (Metadata Tagging)
CREATE TABLE keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL UNIQUE
);

-- Table: keywordreference (Many-to-Many Relationship between documents and keywords)
CREATE TABLE keywordreference (
    article_id INT REFERENCES documents(id) ON DELETE CASCADE,
    keyword_id INT REFERENCES keywords(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, keyword_id)
);

