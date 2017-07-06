DROP TABLE IF EXISTS Jobs;
DROP TABLE IF EXISTS Dimensions;

CREATE TABLE Jobs(
 instanceId SERIAL PRIMARY KEY,
 job JSONB NOT NULL
);

CREATE TABLE Dimensions(
  id SERIAL PRIMARY KEY,
  instanceId INT,
  nodeName TEXT,
  value TEXT,
  nodeId TEXT
);