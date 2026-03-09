'use strict';

const pool = require('../db');

function formatAttendance(row) {
  const record = {
    student_id: row.student_id,
    lesson_id:  row.lesson_id,
    present:    row.present === 1 || row.present === true,
  };
  if (row.student_name != null)  record.student_name  = row.student_name;
  if (row.lesson_date  != null)  record.lesson_date   = row.lesson_date instanceof Date
    ? row.lesson_date.toISOString().slice(0, 10)
    : row.lesson_date;
  if (row.subject_name != null)  record.subject_name  = row.subject_name;
  return record;
}

// PUT /api/attendance
async function markAttendance(req, res, next) {
  try {
    const { student_id, lesson_id, present } = req.body;
    if (student_id == null || lesson_id == null || typeof present !== 'boolean') {
      const err = new Error('student_id, lesson_id and present (boolean) are required');
      err.status = 400;
      return next(err);
    }
    await pool.execute(
      `INSERT INTO attendance (student_id, lesson_id, present)
       VALUES (?, ?, ?)
       ON DUPLICATE KEY UPDATE present = VALUES(present)`,
      [student_id, lesson_id, present]
    );
    const [rows] = await pool.execute(
      'SELECT student_id, lesson_id, present FROM attendance WHERE student_id = ? AND lesson_id = ?',
      [student_id, lesson_id]
    );
    res.json(formatAttendance(rows[0]));
  } catch (err) {
    next(err);
  }
}

// GET /api/attendance/lesson/:lessonId
async function getAttendanceForLesson(req, res, next) {
  try {
    const [rows] = await pool.execute(
      `SELECT a.student_id, a.lesson_id, a.present, s.name AS student_name
       FROM   attendance a
       JOIN   students   s ON s.id = a.student_id
       WHERE  a.lesson_id = ?
       ORDER BY s.name`,
      [req.params.lessonId]
    );
    res.json(rows.map(formatAttendance));
  } catch (err) {
    next(err);
  }
}

// GET /api/attendance/student/:studentId
async function getAttendanceForStudent(req, res, next) {
  try {
    const [rows] = await pool.execute(
      `SELECT a.student_id, a.lesson_id, a.present,
              l.date AS lesson_date, sub.name AS subject_name
       FROM   attendance  a
       JOIN   lessons     l   ON l.id  = a.lesson_id
       JOIN   subjects    sub ON sub.id = l.subject_id
       WHERE  a.student_id = ?
       ORDER BY l.date`,
      [req.params.studentId]
    );
    res.json(rows.map(formatAttendance));
  } catch (err) {
    next(err);
  }
}

module.exports = { markAttendance, getAttendanceForLesson, getAttendanceForStudent };
