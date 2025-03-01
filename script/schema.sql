-- Table: documents (Main Storage Table)
CREATE TABLE documents (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    file_name VARCHAR(255) UNIQUE NOT NULL,  -- Ensure unique file names
    title TEXT NOT NULL,                     -- Title of the document
    abstract TEXT,                            -- Summary
    category VARCHAR(100),                    -- Classification category
    publish_date DATE,                        -- When the file was published
    source VARCHAR(255),                      -- Where it came from
    region_id INT REFERENCES regions(id) ON DELETE SET NULL, -- Links to regions

    s3_bucket VARCHAR(255) NOT NULL,          -- S3 Bucket name
    s3_key VARCHAR(1024) UNIQUE NOT NULL,     -- Full path in S3
    pdf_link VARCHAR(1024),                   -- Public access link (if available)

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Automatically sets timestamp
    last_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP -- Updates when modified
    deleted_at TIMESTAMP NULL -- Soft delete
);

-- Indexes for faster lookups
CREATE INDEX idx_documents_category ON documents(category);
CREATE INDEX idx_documents_publish_date ON documents(publish_date);

-- Table: regions (Geographic Data)
CREATE TABLE regions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Index for faster region lookups
CREATE INDEX idx_regions_name ON regions(name);

-- Table: authors (Tracks file authors)
CREATE TABLE authors (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Table: doc_authors (Many-to-Many Relationship between documents and authors)
CREATE TABLE doc_authors (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    doc_id INT REFERENCES documents(id) ON DELETE CASCADE,
    author_id INT REFERENCES authors(id) ON DELETE CASCADE,
    UNIQUE (doc_id, author_id) -- Prevent duplicate entries
);

-- Indexes for faster lookups in many-to-many relationships
CREATE INDEX idx_doc_authors_doc_id ON doc_authors(doc_id);
CREATE INDEX idx_doc_authors_author_id ON doc_authors(author_id);

-- Table: keywords (Metadata Tagging)
CREATE TABLE keywords (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    keyword VARCHAR(255) UNIQUE NOT NULL
);

-- Table: doc_keywords (Many-to-Many Relationship between documents and keywords)
CREATE TABLE doc_keywords (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    doc_id INT REFERENCES documents(id) ON DELETE CASCADE,
    keyword_id INT REFERENCES keywords(id) ON DELETE CASCADE,
    UNIQUE (doc_id, keyword_id)
);

-- Indexes for keyword searching
CREATE INDEX idx_doc_keywords_doc_id ON doc_keywords(doc_id);
CREATE INDEX idx_doc_keywords_keyword_id ON doc_keywords(keyword_id);