DROP TRIGGER IF EXISTS trg_content_blocks_updated_at ON content_blocks;
DROP TRIGGER IF EXISTS trg_processing_jobs_updated_at ON processing_jobs;
DROP TRIGGER IF EXISTS trg_media_assets_updated_at ON media_assets;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS content_block_media_links;
DROP TABLE IF EXISTS content_blocks;
DROP TABLE IF EXISTS processing_jobs;
DROP TABLE IF EXISTS media_assets;