// Jest setup: polyfills missing globals for JSDOM/whatwg-url
// TextEncoder/TextDecoder are provided by Node's util in many versions
try {
    const util = require('util');
    if (typeof global.TextEncoder === 'undefined' && util.TextEncoder) {
        global.TextEncoder = util.TextEncoder;
    }
    if (typeof global.TextDecoder === 'undefined' && util.TextDecoder) {
        global.TextDecoder = util.TextDecoder;
    }
} catch (e) {
    // ignore if util is unavailable
}

// Ensure crypto.subtle exists for libraries that expect Web Crypto (optional)
if (typeof global.crypto === 'undefined') {
    try {
        const { webcrypto } = require('crypto');
        global.crypto = webcrypto || {};
    } catch (e) {
        global.crypto = {};
    }
}

// Fallback minimal TextEncoder/TextDecoder implementations for older Node versions
if (typeof global.TextEncoder === 'undefined') {
    class TextEncoderPolyfill {
        encode(str = '') {
            // Buffer.from produces a Node Buffer, convert to Uint8Array
            return new Uint8Array(Buffer.from(String(str), 'utf8'));
        }
    }
    global.TextEncoder = TextEncoderPolyfill;
}

if (typeof global.TextDecoder === 'undefined') {
    class TextDecoderPolyfill {
        constructor(encoding = 'utf-8') { this.encoding = encoding; }
        decode(buf) {
            if (buf instanceof Uint8Array || Array.isArray(buf)) {
                return Buffer.from(buf).toString('utf8');
            }
            return String(buf);
        }
    }
    global.TextDecoder = TextDecoderPolyfill;
}

// Mark that we're running under Jest so the app can short-circuit DOM-heavy
// initialization during unit/integration tests.
global.__JEST__ = true;
