'use strict';

const pool = require('./index');

async function initDb() {
  await pool.execute(`
    CREATE TABLE IF NOT EXISTS users (
      id        INT           NOT NULL AUTO_INCREMENT,
      name      VARCHAR(255)  NOT NULL,
      birthday  DATE          NOT NULL,
      PRIMARY KEY (id)
    )
  `);
}

module.exports = { initDb };
