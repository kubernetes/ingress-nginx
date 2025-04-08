const crypto = require('crypto');

function cache_key(req) {
    return crypto.createHash('sha1').update(req.variables.tmp_cache_key).digest('base64');
}

export default { cache_key };
