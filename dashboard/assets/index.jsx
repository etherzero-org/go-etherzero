// Copyright 2017 The go-ethzero Authors
// This file is part of the go-ethzero library.
//
// The go-ethzero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethzero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethzero library. If not, see <http://www.gnu.org/licenses/>.

import React from 'react';
import {hydrate} from 'react-dom';
import {createMuiTheme, MuiThemeProvider} from 'material-ui/styles';

import Dashboard from './components/Dashboard.jsx';

// Theme for the dashboard.
const theme = createMuiTheme({
    palette: {
        type: 'dark',
    },
});

// Renders the whole dashboard.
hydrate(
    <MuiThemeProvider theme={theme}>
        <Dashboard />
    </MuiThemeProvider>,
    document.getElementById('dashboard')
);
