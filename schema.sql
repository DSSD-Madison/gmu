--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4 (Debian 17.4-1.pgdg120+2)
-- Dumped by pg_dump version 17.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: authors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.authors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: doc_authors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.doc_authors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_id uuid,
    author_id uuid
);


--
-- Name: doc_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.doc_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_id uuid,
    category_id uuid
);


--
-- Name: doc_keywords; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.doc_keywords (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_id uuid,
    keyword_id uuid
);


--
-- Name: doc_regions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.doc_regions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_id uuid,
    region_id uuid
);


--
-- Name: documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    file_name character varying(255) NOT NULL,
    title text NOT NULL,
    abstract text,
    publish_date date,
    source character varying(255),
    to_index boolean DEFAULT true,
    s3_file character varying(1024) NOT NULL,
    s3_file_preview character varying(1024),
    pdf_link character varying(1024),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    to_delete boolean DEFAULT false NOT NULL,
    to_generate_preview boolean DEFAULT false
);


--
-- Name: flyway_schema_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.flyway_schema_history (
    installed_rank integer NOT NULL,
    version character varying(50),
    description character varying(200) NOT NULL,
    type character varying(20) NOT NULL,
    script character varying(1000) NOT NULL,
    checksum integer,
    installed_by character varying(100) NOT NULL,
    installed_on timestamp without time zone DEFAULT now() NOT NULL,
    execution_time integer NOT NULL,
    success boolean NOT NULL
);


--
-- Name: keywords; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.keywords (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: regions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.regions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username text NOT NULL,
    password_hash text NOT NULL,
    is_master boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: authors authors_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_name_key UNIQUE (name);


--
-- Name: authors authors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_pkey PRIMARY KEY (id);


--
-- Name: categories categories_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_name_key UNIQUE (name);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: doc_authors doc_authors_doc_id_author_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_doc_id_author_id_key UNIQUE (doc_id, author_id);


--
-- Name: doc_authors doc_authors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_pkey PRIMARY KEY (id);


--
-- Name: doc_categories doc_categories_doc_id_category_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_categories
    ADD CONSTRAINT doc_categories_doc_id_category_id_key UNIQUE (doc_id, category_id);


--
-- Name: doc_categories doc_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_categories
    ADD CONSTRAINT doc_categories_pkey PRIMARY KEY (id);


--
-- Name: doc_keywords doc_keywords_doc_id_keyword_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_doc_id_keyword_id_key UNIQUE (doc_id, keyword_id);


--
-- Name: doc_keywords doc_keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_pkey PRIMARY KEY (id);


--
-- Name: doc_regions doc_regions_doc_id_region_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_regions
    ADD CONSTRAINT doc_regions_doc_id_region_id_key UNIQUE (doc_id, region_id);


--
-- Name: doc_regions doc_regions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_regions
    ADD CONSTRAINT doc_regions_pkey PRIMARY KEY (id);


--
-- Name: documents documents_file_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_file_name_key UNIQUE (file_name);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: documents documents_s3_file_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_s3_file_key UNIQUE (s3_file);


--
-- Name: documents documents_s3_file_preview_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_s3_file_preview_key UNIQUE (s3_file_preview);


--
-- Name: flyway_schema_history flyway_schema_history_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.flyway_schema_history
    ADD CONSTRAINT flyway_schema_history_pk PRIMARY KEY (installed_rank);


--
-- Name: keywords keywords_keyword_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keywords
    ADD CONSTRAINT keywords_keyword_key UNIQUE (name);


--
-- Name: keywords keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.keywords
    ADD CONSTRAINT keywords_pkey PRIMARY KEY (id);


--
-- Name: regions regions_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_name_key UNIQUE (name);


--
-- Name: regions regions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: flyway_schema_history_s_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX flyway_schema_history_s_idx ON public.flyway_schema_history USING btree (success);


--
-- Name: idx_categories_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_categories_name ON public.categories USING btree (name);


--
-- Name: idx_doc_authors_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_authors_author_id ON public.doc_authors USING btree (author_id);


--
-- Name: idx_doc_authors_doc_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_authors_doc_id ON public.doc_authors USING btree (doc_id);


--
-- Name: idx_doc_categories_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_categories_category_id ON public.doc_categories USING btree (category_id);


--
-- Name: idx_doc_categories_doc_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_categories_doc_id ON public.doc_categories USING btree (doc_id);


--
-- Name: idx_doc_keywords_doc_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_keywords_doc_id ON public.doc_keywords USING btree (doc_id);


--
-- Name: idx_doc_keywords_keyword_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_keywords_keyword_id ON public.doc_keywords USING btree (keyword_id);


--
-- Name: idx_doc_regions_doc_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_regions_doc_id ON public.doc_regions USING btree (doc_id);


--
-- Name: idx_doc_regions_region_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_doc_regions_region_id ON public.doc_regions USING btree (region_id);


--
-- Name: idx_documents_publish_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_documents_publish_date ON public.documents USING btree (publish_date);


--
-- Name: idx_regions_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_regions_name ON public.regions USING btree (name);


--
-- Name: doc_authors doc_authors_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.authors(id) ON DELETE CASCADE;


--
-- Name: doc_authors doc_authors_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_doc_id_fkey FOREIGN KEY (doc_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: doc_categories doc_categories_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_categories
    ADD CONSTRAINT doc_categories_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE;


--
-- Name: doc_categories doc_categories_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_categories
    ADD CONSTRAINT doc_categories_doc_id_fkey FOREIGN KEY (doc_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: doc_keywords doc_keywords_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_doc_id_fkey FOREIGN KEY (doc_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: doc_keywords doc_keywords_keyword_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_keyword_id_fkey FOREIGN KEY (keyword_id) REFERENCES public.keywords(id) ON DELETE CASCADE;


--
-- Name: doc_regions doc_regions_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_regions
    ADD CONSTRAINT doc_regions_doc_id_fkey FOREIGN KEY (doc_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: doc_regions doc_regions_region_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.doc_regions
    ADD CONSTRAINT doc_regions_region_id_fkey FOREIGN KEY (region_id) REFERENCES public.regions(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

