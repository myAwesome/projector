'use strict';

const express = require('express');
const router = express.Router();
const healthRoutes = require('./health');
const userRoutes = require('./users');

router.use('/health', healthRoutes);
router.use('/users', userRoutes);

module.exports = router;
