// Small test harness for unit conversion helpers used by chart.html
// Run with: node pkg/web/static/test/conversion_harness.js

function toF(c){ return (c*9.0/5.0)+32.0 }
function toC(f){ return (f-32.0)*5.0/9.0 }
function inHgToMb(v){ return v*33.8638866667 }
function mbToInHg(v){ return v/33.8638866667 }
function mphToKmh(v){ return v*1.609344 }
function kmhToMph(v){ return v/1.609344 }
function inchesToMm(v){ return v*25.4 }
function mmToInches(v){ return v/25.4 }

function approxEqual(a,b, tol=1e-6){ return Math.abs(a-b) <= tol }

function runTests(){
  const tests = []
  // temperature
  tests.push({name: 'C->F 0', got: toF(0), want:32})
  tests.push({name: 'C->F 100', got: toF(100), want:212})
  tests.push({name: 'F->C 32', got: toC(32), want:0})
  tests.push({name: 'F->C 212', got: toC(212), want:100})
  // pressure
  tests.push({name: 'inHg->mb 29.92', got: inHgToMb(29.92), want:1013.25})
  tests.push({name: 'mb->inHg 1013.25', got: mbToInHg(1013.25), want:29.92})
  // wind
  tests.push({name: 'mph->kmh 10', got: mphToKmh(10), want:16.09344})
  tests.push({name: 'kmh->mph 16.09344', got: kmhToMph(16.09344), want:10})
  // rain
  tests.push({name: 'in->mm 1', got: inchesToMm(1), want:25.4})
  tests.push({name: 'mm->in 25.4', got: mmToInches(25.4), want:1})

  let failed = 0
  for(const t of tests){
    // pressure conversions can have slightly larger acceptable error due to constants
    const tol = (t.name.indexOf('inHg')>=0 || t.name.indexOf('mb->inHg')>=0) ? 0.05 : 1e-3
    if(!approxEqual(t.got, t.want, tol)){
      console.error('FAIL', t.name, 'got=', t.got, 'want=', t.want)
      failed++
    } else {
      console.log('ok ', t.name)
    }
  }
  if(failed>0){
    console.error('Some tests failed:', failed)
    process.exit(2)
  } else {
    console.log('All conversion tests passed')
  }
}

if(require.main === module){ runTests() }
