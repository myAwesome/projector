'use strict';

const express = require('express');
const router = express.Router();
const {
  listSubjects,
  getSubject,
  createSubject,
  updateSubject,
  deleteSubject,
} = require('../controllers/subjectController');

router.get('/',       listSubjects);
router.get('/:id',    getSubject);
router.post('/',      createSubject);
router.put('/:id',    updateSubject);
router.delete('/:id', deleteSubject);

module.exports = router;
