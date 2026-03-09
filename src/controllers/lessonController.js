'use strict';

const pool = require('../db');

function formatLesson(row) {
  return {
    id:         row.id,
    subject_id: row.subject_id,
    date:       row.date instanceof Date
      ? row.date.toISOString().slice(0, 10)
      : row.date,
  };
}

// GET /api/lessons
async function listLessons(req, res, next) {
  try {
    const [rows] = await pool.execute('SELECT id, subject_id, date FROM lessons ORDER BY id');
    res.json(rows.map(formatLesson));
  } catch (err) {
    next(err);
  }
}

// GET /api/lessons/:id
async function getLesson(req, res, next) {
  try {
    const [rows] = await pool.execute(
      'SELECT id, subject_id, date FROM lessons WHERE id = ?',
      [req.params.id]
    );
    if (rows.length === 0) {
      const err = new Error('Lesson not found');
      err.status = 404;
      return next(err);
    }
    res.json(formatLesson(rows[0]));
  } catch (err) {
    next(err);
  }
}

// POST /api/lessons
async function createLesson(req, res, next) {
  try {
    const { subject_id, date } = req.body;
    if (!subject_id || !date) {
      const err = new Error('subject_id and date are required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'INSERT INTO lessons (subject_id, date) VALUES (?, ?)',
      [subject_id, date]
    );
    const [rows] = await pool.execute(
      'SELECT id, subject_id, date FROM lessons WHERE id = ?',
      [result.insertId]
    );
    res.status(201).json(formatLesson(rows[0]));
  } catch (err) {
    next(err);
  }
}

// PUT /api/lessons/:id
async function updateLesson(req, res, next) {
  try {
    const { subject_id, date } = req.body;
    if (!subject_id || !date) {
      const err = new Error('subject_id and date are required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'UPDATE lessons SET subject_id = ?, date = ? WHERE id = ?',
      [subject_id, date, req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('Lesson not found');
      err.status = 404;
      return next(err);
    }
    const [rows] = await pool.execute(
      'SELECT id, subject_id, date FROM lessons WHERE id = ?',
      [req.params.id]
    );
    res.json(formatLesson(rows[0]));
  } catch (err) {
    next(err);
  }
}

// DELETE /api/lessons/:id
async function deleteLesson(req, res, next) {
  try {
    const [result] = await pool.execute(
      'DELETE FROM lessons WHERE id = ?',
      [req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('Lesson not found');
      err.status = 404;
      return next(err);
    }
    res.status(204).send();
  } catch (err) {
    next(err);
  }
}

module.exports = { listLessons, getLesson, createLesson, updateLesson, deleteLesson };
