// Minimal bridge to jsondiffpatch for test comparisons.
// Requires `npm install jsondiffpatch` in the project or globally resolvable.
const jsondiffpatch = require('jsondiffpatch').create({ textDiff: { minLength: 10000 } });

const argv = process.argv;
if (argv.length > 3) {
  const j1 = JSON.parse(argv[2]);
  const j2 = JSON.parse(argv[3]);
  const d = jsondiffpatch.diff(j1, j2);
  process.stdout.write(JSON.stringify(d));
  process.exit(0);
} else {
  process.stderr.write('usage: node js/test_helper.js <json1> <json2>\n');
  process.exit(1);
}

