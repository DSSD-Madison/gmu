BEGIN;

-- ========================================
-- Deduplicate KEYWORDS (favor lowercase + normalized)
-- ========================================

-- Step 1 & 2: Normalize and deduplicate
WITH normalized_keywords AS (
  SELECT id, name,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(name, '[:_-]+', ' ', 'g'),
          '\s+', ' ', 'g'
        )
      )
    ) AS norm_name
  FROM keywords
),
grouped_keywords AS (
  SELECT norm_name, MIN(id::text)::uuid AS canonical_id
  FROM normalized_keywords
  GROUP BY norm_name
),
keyword_dupes AS (
  SELECT nk.id AS old_id, gk.canonical_id, nk.norm_name
  FROM normalized_keywords nk
  JOIN grouped_keywords gk ON nk.norm_name = gk.norm_name
  WHERE nk.id != gk.canonical_id
)
-- Repoint doc_keywords to canonical ID
UPDATE doc_keywords dk
SET keyword_id = kd.canonical_id
FROM keyword_dupes kd
WHERE dk.keyword_id = kd.old_id;

-- Step 3: Delete duplicate keyword rows
WITH normalized_keywords AS (
  SELECT id, name,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(name, '[:_-]+', ' ', 'g'),
          '\s+', ' ', 'g'
        )
      )
    ) AS norm_name
  FROM keywords
),
grouped_keywords AS (
  SELECT norm_name, MIN(id::text)::uuid AS canonical_id
  FROM normalized_keywords
  GROUP BY norm_name
),
keyword_dupes AS (
  SELECT nk.id AS old_id
  FROM normalized_keywords nk
  JOIN grouped_keywords gk ON nk.norm_name = gk.norm_name
  WHERE nk.id != gk.canonical_id
)
DELETE FROM keywords
WHERE id IN (SELECT old_id FROM keyword_dupes);

-- Step 4: Rename surviving entries to their normalized form
WITH normalized_keywords AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(name, '[:_-]+', ' ', 'g'),
          '\s+', ' ', 'g'
        )
      )
    ) AS norm_name
  FROM keywords
)
UPDATE keywords k
SET name = nk.norm_name
FROM normalized_keywords nk
WHERE k.id = nk.id AND k.name IS DISTINCT FROM nk.norm_name;


-- ========================================
-- Same process for REGIONS (favor Uppercase)
-- ========================================

-- Step 1 & 2: Normalize and deduplicate
WITH normalized_regions AS (
  SELECT id, name,
    INITCAP(
        TRIM(
            REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[:_-]+', ' ', 'g'),
            '\s+', ' ', 'g'
            )
        )
    ) AS norm_name
  FROM regions
),
grouped_regions AS (
  SELECT norm_name, MIN(id::text)::uuid AS canonical_id
  FROM normalized_regions
  GROUP BY norm_name
),
region_dupes AS (
  SELECT nr.id AS old_id, gr.canonical_id
  FROM normalized_regions nr
  JOIN grouped_regions gr ON nr.norm_name = gr.norm_name
  WHERE nr.id != gr.canonical_id
)
UPDATE doc_regions dr
SET region_id = rd.canonical_id
FROM region_dupes rd
WHERE dr.region_id = rd.old_id;

-- Step 3: Delete duplicate region rows
WITH normalized_regions AS (
  SELECT id, name,
    INITCAP(
        TRIM(
            REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[:_-]+', ' ', 'g'),
            '\s+', ' ', 'g'
            )
        )
    ) AS norm_name
  FROM regions
),
grouped_regions AS (
  SELECT norm_name, MIN(id::text)::uuid AS canonical_id
  FROM normalized_regions
  GROUP BY norm_name
),
region_dupes AS (
  SELECT nr.id AS old_id
  FROM normalized_regions nr
  JOIN grouped_regions gr ON nr.norm_name = gr.norm_name
  WHERE nr.id != gr.canonical_id
)
DELETE FROM regions
WHERE id IN (SELECT old_id FROM region_dupes);

-- Step 4: Rename surviving entries to their normalized form
WITH normalized_regions AS (
  SELECT id,
    INITCAP(
        TRIM(
            REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[:_-]+', ' ', 'g'),
            '\s+', ' ', 'g'
            )
        )
    ) AS norm_name
  FROM regions
)
UPDATE regions r
SET name = nr.norm_name
FROM normalized_regions nr
WHERE r.id = nr.id AND r.name IS DISTINCT FROM nr.norm_name;

COMMIT;
