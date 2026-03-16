DO
$$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_database
      WHERE datname = 'go_db_exercise'
   ) THEN
      CREATE DATABASE go_db_exercise;
   END IF;
END
$$;