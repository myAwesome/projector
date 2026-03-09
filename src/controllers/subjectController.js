'use strict';

const pool = require('../db');

function formatSubject(row) {
  return { id: row.id, name: row.name };
}

// GET /api/subjects
async function listSubjects(req, res, next) {
  try {
    const [rows] = await pool.execute('SELECT id, name FROM subjects ORDER BY id');
    res.json(rows.map(formatSubject));
  } catch (err) {
    next(err);
  }
}

// GET /api/subjects/:id
async function getSubject(req, res, next) {
  try {
    const [rows] = await pool.execute(
      'SELECT id, name FROM subjects WHERE id = ?',
      [req.params.id]
    );
    if (rows.length === 0) {
      const err = new Error('Subject not found');
      err.status = 404;
      return next(err);
    }
    res.json(formatSubject(rows[0]));
  } catch (err) {
    next(err);
  }
}

// POST /api/subjects
async function createSubject(req, res, next) {
  try {
    const { name } = req.body;
    if (!name) {
      const err = new Error('name is required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'INSERT INTO subjects (name) VALUES (?)',
      [name]
    );
    const [rows] = await pool.execute(
      'SELECT id, name FROM subjects WHERE id = ?',
      [result.insertId]
    );
    res.status(201).json(formatSubject(rows[0]));
  } catch (err) {
    next(err);
  }
}

// PUT /api/subjects/:id
async function updateSubject(req, res, next) {
  try {
    const { name } = req.body;
    if (!name) {
      const err = new Error('name is required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'UPDATE subjects SET name = ? WHERE id = ?',
      [name, req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('Subject not found');
      err.status = 404;
      return next(err);
    }
    const [rows] = await pool.execute(
      'SELECT id, name FROM subjects WHERE id = ?',
      [req.params.id]
    );
    res.json(formatSubject(rows[0]));
  } catch (err) {
    next(err);
  }
}

// DELETE /api/subjects/:id
async function deleteSubject(req, res, next) {
  try {
    const [result] = await pool.execute(
      'DELETE FROM subjects WHERE id = ?',
      [req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('Subject not found');
      err.status = 404;
      return next(err);
    }
    res.status(204).send();
  } catch (err) {
    next(err);
  }
}

module.exports = { listSubjects, getSubject, createSubject, updateSubject, deleteSubject };
