#!/usr/bin/env bun
import React from 'react';
import { render } from 'ink';
import App from './app.js';
import { loadAll } from './data/loader.js';

const initialData = loadAll();

render(<App initialData={initialData} />);
