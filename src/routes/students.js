'use strict';

const express = require('express');
const router = express.Router();
const {
  listStudents,
  getStudent,
  createStudent,
  updateStudent,
  deleteStudent,
} = require('../controllers/studentController');

router.get('/',       listStudents);
router.get('/:id',    getStudent);
router.post('/',      createStudent);
router.put('/:id',    updateStudent);
router.delete('/:id', deleteStudent);

module.exports = router;
