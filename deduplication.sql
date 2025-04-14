BEGIN;

-- ========================================
-- Deduplicate KEYWORDS (lowercase + remove () + strip punctuation)
-- ========================================

WITH normalized_keywords AS (
  SELECT id, name,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),  -- remove all parentheses
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
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
  SELECT nk.id AS old_id, gk.canonical_id
  FROM normalized_keywords nk
  JOIN grouped_keywords gk ON nk.norm_name = gk.norm_name
  WHERE nk.id != gk.canonical_id
)
UPDATE doc_keywords dk
SET keyword_id = kd.canonical_id
FROM keyword_dupes kd
WHERE dk.keyword_id = kd.old_id;

-- Delete duplicate keywords
WITH normalized_keywords AS (
  SELECT id, name,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
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

-- Rename surviving keyword entries
WITH normalized_keywords AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
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
-- Deduplicate CATEGORIES (lowercase + remove () + strip punctuation)
-- ========================================

WITH normalized_categories AS (
  SELECT id, name,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
        )
      )
    ) AS norm_name
  FROM categories
),
grouped_categories AS (
  SELECT norm_name, MIN(id::text)::uuid AS canonical_id
  FROM normalized_categories
  GROUP BY norm_name
),
category_dupes AS (
  SELECT nc.id AS old_id, gc.canonical_id
  FROM normalized_categories nc
  JOIN grouped_categories gc ON nc.norm_name = gc.norm_name
  WHERE nc.id != gc.canonical_id
)
UPDATE doc_categories dc
SET category_id = cd.canonical_id
FROM category_dupes cd
WHERE dc.category_id = cd.old_id;

-- Delete duplicate categories
WITH normalized_categories AS (
  SELECT id, name,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
        )
      )
    ) AS norm_name
  FROM categories
),
grouped_categories AS (
  SELECT norm_name, MIN(id::text)::uuid AS canonical_id
  FROM normalized_categories
  GROUP BY norm_name
),
category_dupes AS (
  SELECT nc.id AS old_id
  FROM normalized_categories nc
  JOIN grouped_categories gc ON nc.norm_name = gc.norm_name
  WHERE nc.id != gc.canonical_id
)
DELETE FROM categories
WHERE id IN (SELECT old_id FROM category_dupes);

-- Rename surviving category entries
WITH normalized_categories AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
        )
      )
    ) AS norm_name
  FROM categories
)
UPDATE categories c
SET name = nc.norm_name
FROM normalized_categories nc
WHERE c.id = nc.id AND c.name IS DISTINCT FROM nc.norm_name;


-- ========================================
-- Deduplicate REGIONS (title case + remove () + strip punctuation)
-- ========================================

WITH normalized_regions AS (
  SELECT id, name,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
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

-- Delete duplicate regions
WITH normalized_regions AS (
  SELECT id, name,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
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

-- Rename surviving region entries
WITH normalized_regions AS (
  SELECT id,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(name, '[()]', '', 'g'),
            '^[[:punct:]\s]+|[[:punct:]\s]+$', '', 'g'
          ),
          '[:_-]+', ' ', 'g'
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
