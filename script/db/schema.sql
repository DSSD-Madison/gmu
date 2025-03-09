-- Table: regions (Geographic Data)
CREATE TABLE IF NOT EXISTS regions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Index for faster region lookups
CREATE INDEX IF NOT EXISTS idx_regions_name ON regions(name);

-- Table: authors (Tracks file authors)
CREATE TABLE IF NOT EXISTS authors (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Table: keywords (Metadata Tagging)
CREATE TABLE IF NOT EXISTS keywords (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    keyword VARCHAR(255) UNIQUE NOT NULL
);

-- Table: documents (Main Storage Table)
CREATE TABLE IF NOT EXISTS documents (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    file_name VARCHAR(255) UNIQUE NOT NULL,  -- Ensure unique file names
    title TEXT NOT NULL,                     -- Title of the document
    abstract TEXT,                            -- Summary
    category VARCHAR(100),                    -- Classification category
    publish_date DATE,                        -- When the file was published
    source VARCHAR(255),                      -- Where it came from
    region_id UUID REFERENCES regions(id) ON DELETE SET NULL, -- Links to regions

    s3_file VARCHAR(1024) UNIQUE NOT NULL,    -- Full path in S3
    s3_file_preview VARCHAR(1024) UNIQUE,     -- Path in S3
    pdf_link VARCHAR(1024),                   -- Public access link (if available)

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Automatically sets timestamp
    deleted_at TIMESTAMP NULL -- Soft delete
);

-- Indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_documents_category ON documents(category);
CREATE INDEX IF NOT EXISTS idx_documents_publish_date ON documents(publish_date);

-- Table: doc_authors (Many-to-Many Relationship between documents and authors)
CREATE TABLE IF NOT EXISTS doc_authors (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    doc_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    author_id UUID REFERENCES authors(id) ON DELETE CASCADE,
    UNIQUE (doc_id, author_id) -- Prevent duplicate entries
);

-- Indexes for faster lookups in many-to-many relationships
CREATE INDEX IF NOT EXISTS idx_doc_authors_doc_id ON doc_authors(doc_id);
CREATE INDEX IF NOT EXISTS idx_doc_authors_author_id ON doc_authors(author_id);

-- Table: doc_keywords (Many-to-Many Relationship between documents and keywords)
CREATE TABLE IF NOT EXISTS doc_keywords (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    doc_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    keyword_id UUID REFERENCES keywords(id) ON DELETE CASCADE,
    UNIQUE (doc_id, keyword_id)
);

-- Indexes for keyword searching
CREATE INDEX IF NOT EXISTS idx_doc_keywords_doc_id ON doc_keywords(doc_id);
CREATE INDEX IF NOT EXISTS idx_doc_keywords_keyword_id ON doc_keywords(keyword_id);