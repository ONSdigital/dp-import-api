DROP TABLE IF EXISTS Jobs;
DROP TABLE IF EXISTS Instances;
DROP TABLE IF EXISTS Dimensions;

CREATE TABLE Jobs(
  jobId TEXT PRIMARY KEY,
  job JSONB NOT NULL
);

CREATE TABLE Instances(
 instanceId TEXT PRIMARY KEY,
 jobId TEXT,
 instance JSONB NOT NULL
);

CREATE TABLE Dimensions(
  id SERIAL PRIMARY KEY,
  instanceId TEXT,
  nodeName TEXT,
  value TEXT,
  nodeId TEXT
);