DROP TABLE IF EXISTS Jobs;
DROP TABLE IF EXISTS Instances;
DROP TABLE IF EXISTS Dimensions;

CREATE TABLE Jobs(
  jobId SERIAL PRIMARY KEY,
  job JSONB NOT NULL
);

CREATE TABLE Instances(
 instanceId SERIAL PRIMARY KEY,
 jobId INT,
 instance JSONB NOT NULL
);

CREATE TABLE Dimensions(
  id SERIAL PRIMARY KEY,
  instanceId INT,
  nodeName TEXT,
  value TEXT,
  nodeId TEXT
);