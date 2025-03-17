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
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: authors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.authors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.authors OWNER TO postgres;

--
-- Name: doc_authors; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.doc_authors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_id uuid,
    author_id uuid
);


ALTER TABLE public.doc_authors OWNER TO postgres;

--
-- Name: doc_keywords; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.doc_keywords (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    doc_id uuid,
    keyword_id uuid
);


ALTER TABLE public.doc_keywords OWNER TO postgres;

--
-- Name: documents; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.documents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    file_name character varying(255) NOT NULL,
    title text NOT NULL,
    abstract text,
    category character varying(100),
    publish_date date,
    source character varying(255),
    region_id uuid,
    s3_file character varying(1024) NOT NULL,
    s3_file_preview character varying(1024),
    pdf_link character varying(1024),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone
);


ALTER TABLE public.documents OWNER TO postgres;

--
-- Name: keywords; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.keywords (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    keyword character varying(255) NOT NULL
);


ALTER TABLE public.keywords OWNER TO postgres;

--
-- Name: regions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.regions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL
);


ALTER TABLE public.regions OWNER TO postgres;

--
-- Name: authors authors_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_name_key UNIQUE (name);


--
-- Name: authors authors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_pkey PRIMARY KEY (id);


--
-- Name: doc_authors doc_authors_doc_id_author_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_doc_id_author_id_key UNIQUE (doc_id, author_id);


--
-- Name: doc_authors doc_authors_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_pkey PRIMARY KEY (id);


--
-- Name: doc_keywords doc_keywords_doc_id_keyword_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_doc_id_keyword_id_key UNIQUE (doc_id, keyword_id);


--
-- Name: doc_keywords doc_keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_pkey PRIMARY KEY (id);


--
-- Name: documents documents_file_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_file_name_key UNIQUE (file_name);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: documents documents_s3_file_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_s3_file_key UNIQUE (s3_file);


--
-- Name: documents documents_s3_file_preview_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_s3_file_preview_key UNIQUE (s3_file_preview);


--
-- Name: keywords keywords_keyword_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.keywords
    ADD CONSTRAINT keywords_keyword_key UNIQUE (keyword);


--
-- Name: keywords keywords_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.keywords
    ADD CONSTRAINT keywords_pkey PRIMARY KEY (id);


--
-- Name: regions regions_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_name_key UNIQUE (name);


--
-- Name: regions regions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.regions
    ADD CONSTRAINT regions_pkey PRIMARY KEY (id);


--
-- Name: idx_doc_authors_author_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_doc_authors_author_id ON public.doc_authors USING btree (author_id);


--
-- Name: idx_doc_authors_doc_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_doc_authors_doc_id ON public.doc_authors USING btree (doc_id);


--
-- Name: idx_doc_keywords_doc_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_doc_keywords_doc_id ON public.doc_keywords USING btree (doc_id);


--
-- Name: idx_doc_keywords_keyword_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_doc_keywords_keyword_id ON public.doc_keywords USING btree (keyword_id);


--
-- Name: idx_documents_category; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_documents_category ON public.documents USING btree (category);


--
-- Name: idx_documents_publish_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_documents_publish_date ON public.documents USING btree (publish_date);


--
-- Name: idx_regions_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_regions_name ON public.regions USING btree (name);


--
-- Name: doc_authors doc_authors_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.authors(id) ON DELETE CASCADE;


--
-- Name: doc_authors doc_authors_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_authors
    ADD CONSTRAINT doc_authors_doc_id_fkey FOREIGN KEY (doc_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: doc_keywords doc_keywords_doc_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_doc_id_fkey FOREIGN KEY (doc_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- Name: doc_keywords doc_keywords_keyword_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.doc_keywords
    ADD CONSTRAINT doc_keywords_keyword_id_fkey FOREIGN KEY (keyword_id) REFERENCES public.keywords(id) ON DELETE CASCADE;


--
-- Name: documents documents_region_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_region_id_fkey FOREIGN KEY (region_id) REFERENCES public.regions(id) ON DELETE SET NULL;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

