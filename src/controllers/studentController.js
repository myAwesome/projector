'use strict';

const pool = require('../db');

function formatStudent(row) {
  return { id: row.id, name: row.name };
}

// GET /api/students
async function listStudents(req, res, next) {
  try {
    const [rows] = await pool.execute('SELECT id, name FROM students ORDER BY id');
    res.json(rows.map(formatStudent));
  } catch (err) {
    next(err);
  }
}

// GET /api/students/:id
async function getStudent(req, res, next) {
  try {
    const [rows] = await pool.execute(
      'SELECT id, name FROM students WHERE id = ?',
      [req.params.id]
    );
    if (rows.length === 0) {
      const err = new Error('Student not found');
      err.status = 404;
      return next(err);
    }
    res.json(formatStudent(rows[0]));
  } catch (err) {
    next(err);
  }
}

// POST /api/students
async function createStudent(req, res, next) {
  try {
    const { name } = req.body;
    if (!name) {
      const err = new Error('name is required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'INSERT INTO students (name) VALUES (?)',
      [name]
    );
    const [rows] = await pool.execute(
      'SELECT id, name FROM students WHERE id = ?',
      [result.insertId]
    );
    res.status(201).json(formatStudent(rows[0]));
  } catch (err) {
    next(err);
  }
}

// PUT /api/students/:id
async function updateStudent(req, res, next) {
  try {
    const { name } = req.body;
    if (!name) {
      const err = new Error('name is required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'UPDATE students SET name = ? WHERE id = ?',
      [name, req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('Student not found');
      err.status = 404;
      return next(err);
    }
    const [rows] = await pool.execute(
      'SELECT id, name FROM students WHERE id = ?',
      [req.params.id]
    );
    res.json(formatStudent(rows[0]));
  } catch (err) {
    next(err);
  }
}

// DELETE /api/students/:id
async function deleteStudent(req, res, next) {
  try {
    const [result] = await pool.execute(
      'DELETE FROM students WHERE id = ?',
      [req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('Student not found');
      err.status = 404;
      return next(err);
    }
    res.status(204).send();
  } catch (err) {
    next(err);
  }
}

module.exports = { listStudents, getStudent, createStudent, updateStudent, deleteStudent };
