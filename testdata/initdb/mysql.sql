DROP TABLE IF EXISTS various_types;

CREATE TABLE various_types (
  col_int INT NOT NULL,
  col_tinyint TINYINT NOT NULL,
  col_smallint SMALLINT NOT NULL,
  col_mediumint MEDIUMINT NOT NULL,
  col_bigint BIGINT NOT NULL,
  col_char CHAR (50) NOT NULL,
  col_varchar VARCHAR (50) NOT NULL,
  col_text TEXT NOT NULL,
  col_boolean BOOLEAN NOT NULL,
  col_date DATE NOT NULL,
  col_time TIME NOT NULL,
  col_timestamp TIMESTAMP NOT NULL,
  col_datetime DATETIME NOT NULL,
  col_enum ENUM('ONE', 'TWO', 'THREE') NOT NULL
);

INSERT INTO various_types (
  col_int,
  col_tinyint,
  col_smallint,
  col_mediumint,
  col_bigint,
  col_char,
  col_varchar,
  col_text,
  col_boolean,
  col_date,
  col_time,
  col_timestamp,
  col_datetime,
  col_enum
) VALUES (
  1,
  2,
  3,
  4,
  5,
  'this is char',
  'this is varchar',
  'this is text',
  true,
  '2022-01-02',
  '09:56:59',
  '2022-01-02 09:56:59',
  '2022-01-02 09:56:59',
  'TWO'
);
