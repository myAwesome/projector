'use strict';

const express = require('express');
const router = express.Router();
const healthRoutes     = require('./health');
const userRoutes       = require('./users');
const studentRoutes    = require('./students');
const subjectRoutes    = require('./subjects');
const lessonRoutes     = require('./lessons');
const attendanceRoutes = require('./attendance');

router.use('/health',     healthRoutes);
router.use('/users',      userRoutes);
router.use('/students',   studentRoutes);
router.use('/subjects',   subjectRoutes);
router.use('/lessons',    lessonRoutes);
router.use('/attendance', attendanceRoutes);

module.exports = router;
