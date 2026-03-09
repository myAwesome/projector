'use strict';

const express = require('express');
const router = express.Router();
const {
  listLessons,
  getLesson,
  createLesson,
  updateLesson,
  deleteLesson,
} = require('../controllers/lessonController');

router.get('/',       listLessons);
router.get('/:id',    getLesson);
router.post('/',      createLesson);
router.put('/:id',    updateLesson);
router.delete('/:id', deleteLesson);

module.exports = router;
