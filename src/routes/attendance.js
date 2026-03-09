'use strict';

const express = require('express');
const router = express.Router();
const {
  markAttendance,
  getAttendanceForLesson,
  getAttendanceForStudent,
} = require('../controllers/attendanceController');

router.put('/',                        markAttendance);
router.get('/lesson/:lessonId',        getAttendanceForLesson);
router.get('/student/:studentId',      getAttendanceForStudent);

module.exports = router;
