'use strict';

const pool = require('../db');

// mysql2 returns DATE columns as JS Date objects. Format them as YYYY-MM-DD.
function formatUser(row) {
  return {
    id:       row.id,
    name:     row.name,
    birthday: row.birthday instanceof Date
      ? row.birthday.toISOString().slice(0, 10)
      : row.birthday,
  };
}

// GET /api/users
async function listUsers(req, res, next) {
  try {
    const [rows] = await pool.execute('SELECT id, name, birthday FROM users ORDER BY id');
    res.json(rows.map(formatUser));
  } catch (err) {
    next(err);
  }
}

// GET /api/users/:id
async function getUser(req, res, next) {
  try {
    const [rows] = await pool.execute(
      'SELECT id, name, birthday FROM users WHERE id = ?',
      [req.params.id]
    );
    if (rows.length === 0) {
      const err = new Error('User not found');
      err.status = 404;
      return next(err);
    }
    res.json(formatUser(rows[0]));
  } catch (err) {
    next(err);
  }
}

// POST /api/users
async function createUser(req, res, next) {
  try {
    const { name, birthday } = req.body;
    if (!name || !birthday) {
      const err = new Error('name and birthday are required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'INSERT INTO users (name, birthday) VALUES (?, ?)',
      [name, birthday]
    );
    const [rows] = await pool.execute(
      'SELECT id, name, birthday FROM users WHERE id = ?',
      [result.insertId]
    );
    res.status(201).json(formatUser(rows[0]));
  } catch (err) {
    next(err);
  }
}

// PUT /api/users/:id
async function updateUser(req, res, next) {
  try {
    const { name, birthday } = req.body;
    if (!name || !birthday) {
      const err = new Error('name and birthday are required');
      err.status = 400;
      return next(err);
    }
    const [result] = await pool.execute(
      'UPDATE users SET name = ?, birthday = ? WHERE id = ?',
      [name, birthday, req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('User not found');
      err.status = 404;
      return next(err);
    }
    const [rows] = await pool.execute(
      'SELECT id, name, birthday FROM users WHERE id = ?',
      [req.params.id]
    );
    res.json(formatUser(rows[0]));
  } catch (err) {
    next(err);
  }
}

// DELETE /api/users/:id
async function deleteUser(req, res, next) {
  try {
    const [result] = await pool.execute(
      'DELETE FROM users WHERE id = ?',
      [req.params.id]
    );
    if (result.affectedRows === 0) {
      const err = new Error('User not found');
      err.status = 404;
      return next(err);
    }
    res.status(204).send();
  } catch (err) {
    next(err);
  }
}

module.exports = { listUsers, getUser, createUser, updateUser, deleteUser };
