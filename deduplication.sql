BEGIN;

-- STEP 1: Delete (doc_id, keyword_id) rows that would cause conflicts after update
WITH normalized_keywords AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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
),
conflicts AS (
  SELECT dk.doc_id, kd.old_id
  FROM doc_keywords dk
  JOIN keyword_dupes kd ON dk.keyword_id = kd.old_id
  WHERE EXISTS (
    SELECT 1
    FROM doc_keywords dk2
    WHERE dk2.doc_id = dk.doc_id AND dk2.keyword_id = kd.canonical_id
  )
)
DELETE FROM doc_keywords
WHERE (doc_id, keyword_id) IN (SELECT doc_id, old_id FROM conflicts);

-- STEP 2: Update keyword_id in doc_keywords to canonical
WITH normalized_keywords AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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

-- STEP 3: Delete duplicate keyword rows
WITH normalized_keywords AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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

-- STEP 4: Normalize surviving keyword names
WITH normalized_keywords AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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
-- Deduplicate CATEGORIES
-- ========================================

-- STEP 1: Delete (doc_id, category_id) rows that would cause conflicts after update
WITH normalized_categories AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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
),
conflicts AS (
  SELECT dc.doc_id, cd.old_id
  FROM doc_categories dc
  JOIN category_dupes cd ON dc.category_id = cd.old_id
  WHERE EXISTS (
    SELECT 1
    FROM doc_categories dc2
    WHERE dc2.doc_id = dc.doc_id AND dc2.category_id = cd.canonical_id
  )
)
DELETE FROM doc_categories
WHERE (doc_id, category_id) IN (SELECT doc_id, old_id FROM conflicts);

-- STEP 2: Update category_id in doc_categories
WITH normalized_categories AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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

-- STEP 3: Delete duplicate category rows
WITH normalized_categories AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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

-- STEP 4: Normalize surviving category names
WITH normalized_categories AS (
  SELECT id,
    LOWER(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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
-- Deduplicate REGIONS
-- ========================================

-- STEP 1: Delete (doc_id, region_id) rows that would cause conflicts after update
WITH normalized_regions AS (
  SELECT id,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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
),
conflicts AS (
  SELECT dr.doc_id, rd.old_id
  FROM doc_regions dr
  JOIN region_dupes rd ON dr.region_id = rd.old_id
  WHERE EXISTS (
    SELECT 1
    FROM doc_regions dr2
    WHERE dr2.doc_id = dr.doc_id AND dr2.region_id = rd.canonical_id
  )
)
DELETE FROM doc_regions
WHERE (doc_id, region_id) IN (SELECT doc_id, old_id FROM conflicts);

-- STEP 2: Update region_id in doc_regions
WITH normalized_regions AS (
  SELECT id,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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

-- STEP 3: Delete duplicate region rows
WITH normalized_regions AS (
  SELECT id,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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

-- STEP 4: Normalize surviving region names
WITH normalized_regions AS (
  SELECT id,
    INITCAP(
      TRIM(
        REGEXP_REPLACE(
          REGEXP_REPLACE(
            REGEXP_REPLACE(
              REGEXP_REPLACE(
                REGEXP_REPLACE(
                  REGEXP_REPLACE(
                    REGEXP_REPLACE(name, '''s\b', '', 'g'),
                    '''', '', 'g'
                  ),
                  '[()]', '', 'g'
                ),
                '[[:punct:]]+', ' ', 'g'
              ),
              '[–—_:\\-]+', ' ', 'g'
            ),
            '\s+', ' ', 'g'
          ),
          '^\s+|\s+$', '', 'g'
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
