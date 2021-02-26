/*

   Copyright 2020,2021 Avi Zimmerman

   This file is part of kvdi.

   kvdi is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   kvdi is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with kvdi.  If not, see <https://www.gnu.org/licenses/>.

*/

/*import { transform as babelTransform } from 'babel-core';
import babelTransformDynamicImport from 'babel-plugin-syntax-dynamic-import';
import babelTransformES2015ModulesSystemJS from 'babel-plugin-transform-es2015-modules-systemjs';*/

// sadly, due to how rollup works, we can't use es6 imports here
var babelTransform = require('babel-core').transform;
var babelTransformDynamicImport = require('babel-plugin-syntax-dynamic-import');
var babelTransformES2015ModulesSystemJS = require('babel-plugin-transform-es2015-modules-systemjs');
var babelPresetES2015 = require('babel-preset-es2015');

self.onmessage = function (evt) {
    // transform source with Babel
    var output = babelTransform(evt.data.source, {
      compact: false,
      filename: evt.data.key + '!transpiled',
      sourceFileName: evt.data.key,
      moduleIds: false,
      sourceMaps: 'inline',
      babelrc: false,
      plugins: [babelTransformDynamicImport, babelTransformES2015ModulesSystemJS],
      presets: [babelPresetES2015],
    });

    self.postMessage({key: evt.data.key, code: output.code, source: evt.data.source});
};
