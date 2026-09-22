slang.run {
    encode zstd gzip
    header {
        X-Content-Type-Options nosniff
        Referrer-Policy same-origin
        X-Frame-Options DENY
        Strict-Transport-Security "max-age=31536000"
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
    }
    handle_path /api/auth/* {
        reverse_proxy 127.0.0.1:8201
    }
    handle_path /api/repo/* {
        reverse_proxy 127.0.0.1:8202
    }
    handle_path /api/meta/* {
        reverse_proxy 127.0.0.1:8203
    }
    handle_path /api/usage/* {
        reverse_proxy 127.0.0.1:8204
    }
    handle_path /api/customer/* {
        reverse_proxy 127.0.0.1:8205
    }
    handle_path /api/narrow/* {
        reverse_proxy 127.0.0.1:8206
    }
    handle_path /api/deployment/* {
        reverse_proxy 127.0.0.1:8207
    }
    handle_path /api/search/* {
        reverse_proxy 127.0.0.1:8208
    }
    handle_path /api/vision/* {
        reverse_proxy 127.0.0.1:8209
    }
    handle /api/* {
        respond "Unknown API" 404
    }
    handle /admin/meta {
        redir /admin/meta/ 308
    }
    handle_path /admin/meta/* {
        root * /srv/slang/meta
        file_server
    }
    handle {
        root * /srv/slang/studio
        try_files {path} /index.html
        header Cache-Control no-cache
        file_server
    }
}

slang.bitspark.com {
    encode zstd gzip
    header {
        X-Content-Type-Options nosniff
        Referrer-Policy same-origin
        Strict-Transport-Security "max-age=31536000"
    }
    @legacy_app path /app /app/*
    redir @legacy_app https://slang.run/ 308
    root * /srv/slang/website/site
    try_files {path} {path}/index.html
    file_server
}

tryslang.com, www.tryslang.com, playground.tryslang.com {
    redir https://slang.bitspark.com/ 308
}
