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

  await pool.execute(`
    CREATE TABLE IF NOT EXISTS students (
      id    INT          NOT NULL AUTO_INCREMENT,
      name  VARCHAR(255) NOT NULL,
      PRIMARY KEY (id)
    )
  `);

  await pool.execute(`
    CREATE TABLE IF NOT EXISTS subjects (
      id    INT          NOT NULL AUTO_INCREMENT,
      name  VARCHAR(255) NOT NULL,
      PRIMARY KEY (id)
    )
  `);

  await pool.execute(`
    CREATE TABLE IF NOT EXISTS lessons (
      id         INT  NOT NULL AUTO_INCREMENT,
      subject_id INT  NOT NULL,
      date       DATE NOT NULL,
      PRIMARY KEY (id),
      CONSTRAINT fk_lessons_subject
        FOREIGN KEY (subject_id) REFERENCES subjects (id) ON DELETE CASCADE
    )
  `);

  await pool.execute(`
    CREATE TABLE IF NOT EXISTS attendance (
      student_id INT        NOT NULL,
      lesson_id  INT        NOT NULL,
      present    TINYINT(1) NOT NULL DEFAULT 0,
      PRIMARY KEY (student_id, lesson_id),
      CONSTRAINT fk_attendance_student
        FOREIGN KEY (student_id) REFERENCES students (id) ON DELETE CASCADE,
      CONSTRAINT fk_attendance_lesson
        FOREIGN KEY (lesson_id)  REFERENCES lessons  (id) ON DELETE CASCADE
    )
  `);
}

module.exports = { initDb };
