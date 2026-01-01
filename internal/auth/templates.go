package auth

const setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LINE CLI - Connect Your Account</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #0F0F0F;
            --bg-card: #1A1A1A;
            --bg-input: #252525;
            --bg-elevated: #2A2A2A;
            --bg-hover: #333333;
            --border: #333333;
            --border-focus: #06C755;
            --text: #FFFFFF;
            --text-secondary: #B3B3B3;
            --text-muted: #737373;
            --text-dim: #525252;
            --line-green: #06C755;
            --line-green-dark: #05A847;
            --line-green-light: rgba(6, 199, 85, 0.15);
            --success: #06C755;
            --success-light: rgba(6, 199, 85, 0.15);
            --error: #EF4444;
            --error-light: rgba(239, 68, 68, 0.15);
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        html { height: 100%%; }

        body {
            font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%%;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem 3rem;
            position: relative;
        }

        body::before {
            content: '';
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            background:
                radial-gradient(ellipse at top, rgba(6, 199, 85, 0.08) 0%%, transparent 50%%),
                radial-gradient(ellipse at bottom right, rgba(6, 199, 85, 0.04) 0%%, transparent 40%%);
            pointer-events: none;
        }

        .hidden { display: none !important; }

        .container {
            width: 100%%;
            max-width: 420px;
            position: relative;
            z-index: 1;
            text-align: center;
        }

        /* Logo */
        .logo {
            display: flex;
            justify-content: center;
            margin-bottom: 0.75rem;
            animation: fadeDown 0.4s ease-out;
        }

        .logo svg {
            height: 40px;
            width: auto;
        }

        @keyframes fadeDown {
            from { opacity: 0; transform: translateY(-8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* CLI Badge */
        .badge-wrapper {
            display: flex;
            justify-content: center;
            margin-bottom: 1.25rem;
            animation: fadeDown 0.4s ease-out 0.05s both;
        }

        .cli-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            background: var(--line-green-light);
            color: var(--line-green);
            font-size: 0.6875rem;
            font-weight: 600;
            padding: 0.375rem 0.75rem;
            border-radius: 100px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .cli-badge svg {
            width: 12px;
            height: 12px;
        }

        h1 {
            font-size: 1.5rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 0.25rem;
            text-align: center;
            animation: fadeDown 0.4s ease-out 0.1s both;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.9375rem;
            margin-bottom: 1.5rem;
            text-align: center;
            animation: fadeDown 0.4s ease-out 0.15s both;
        }

        /* Accounts section */
        .accounts-section {
            margin-bottom: 1rem;
            animation: fadeUp 0.4s ease-out 0.2s both;
            text-align: left;
        }

        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 0.75rem;
        }

        .section-title {
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-muted);
        }

        .account-count {
            font-size: 0.6875rem;
            color: var(--text-muted);
            background: var(--bg-input);
            padding: 0.25rem 0.625rem;
            border-radius: 100px;
            border: 1px solid var(--border);
        }

        .accounts-list {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            margin-bottom: 0.75rem;
        }

        .account-card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 10px;
            padding: 0.875rem 1rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
            transition: all 0.2s ease;
        }

        .account-card:hover {
            border-color: #D1D5DB;
            box-shadow: 0 2px 8px rgba(0,0,0,0.04);
        }

        .account-card.primary {
            border-color: rgba(6, 199, 85, 0.3);
            background: linear-gradient(135deg, var(--bg-card), rgba(6, 199, 85, 0.03));
        }

        .account-avatar {
            width: 36px;
            height: 36px;
            background: linear-gradient(135deg, var(--line-green), var(--line-green-dark));
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 700;
            font-size: 14px;
            color: white;
            flex-shrink: 0;
        }

        .account-info { flex: 1; min-width: 0; }

        .account-name {
            font-size: 0.875rem;
            font-weight: 600;
            color: var(--text);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .account-bot {
            font-size: 0.75rem;
            color: var(--text-muted);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }

        .account-actions {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            flex-shrink: 0;
        }

        .primary-badge {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            font-size: 0.625rem;
            font-weight: 600;
            color: var(--line-green);
            background: var(--line-green-light);
            padding: 0.25rem 0.5rem;
            border-radius: 100px;
        }

        .primary-badge svg { width: 10px; height: 10px; }

        .set-primary-btn {
            font-size: 0.6875rem;
            color: var(--text-muted);
            background: var(--bg-input);
            border: 1px solid var(--border);
            padding: 0.25rem 0.625rem;
            border-radius: 100px;
            cursor: pointer;
            transition: all 0.2s ease;
            opacity: 0;
            font-family: inherit;
        }

        .account-card:hover .set-primary-btn { opacity: 1; }

        .set-primary-btn:hover {
            background: var(--border);
            color: var(--text);
        }

        .remove-btn {
            width: 24px;
            height: 24px;
            background: transparent;
            border: none;
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            color: var(--text-muted);
            transition: all 0.2s ease;
            opacity: 0;
        }

        .account-card:hover .remove-btn { opacity: 1; }

        .remove-btn:hover {
            background: var(--error-light);
            color: var(--error);
        }

        .remove-btn svg { width: 14px; height: 14px; }

        .add-account-btn {
            width: 100%%;
            background: transparent;
            border: 1.5px dashed var(--border);
            border-radius: 10px;
            padding: 1rem;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            color: var(--text-muted);
            font-size: 0.8125rem;
            font-weight: 500;
            font-family: inherit;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .add-account-btn:hover {
            border-color: var(--line-green);
            color: var(--line-green);
            background: rgba(6, 199, 85, 0.03);
        }

        .add-account-btn svg { width: 16px; height: 16px; }

        /* Empty state */
        .empty-state {
            text-align: center;
            padding: 2rem 1.5rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            margin-bottom: 1rem;
            animation: fadeUp 0.4s ease-out 0.2s both;
        }

        .empty-state-icon {
            width: 48px;
            height: 48px;
            margin: 0 auto 0.875rem;
            background: var(--line-green-light);
            border-radius: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .empty-state-icon svg {
            width: 24px;
            height: 24px;
            color: var(--line-green);
        }

        .empty-state h3 {
            font-size: 0.9375rem;
            font-weight: 600;
            margin-bottom: 0.25rem;
        }

        .empty-state p {
            font-size: 0.8125rem;
            color: var(--text-muted);
        }

        /* Form card */
        .form-card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            overflow: hidden;
            box-shadow: 0 4px 24px rgba(0, 0, 0, 0.04);
            animation: fadeUp 0.4s ease-out 0.25s both;
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .form-header {
            padding: 1rem 1.25rem;
            border-bottom: 1px solid var(--border);
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .form-header h2 {
            font-size: 0.9375rem;
            font-weight: 600;
        }

        .close-btn {
            width: 28px;
            height: 28px;
            background: var(--bg-input);
            border: 1px solid var(--border);
            border-radius: 8px;
            display: none;
            align-items: center;
            justify-content: center;
            cursor: pointer;
            color: var(--text-muted);
            transition: all 0.2s ease;
        }

        .close-btn.show { display: flex; }

        .close-btn:hover {
            background: var(--border);
            color: var(--text);
        }

        .close-btn svg { width: 14px; height: 14px; }

        .form-body {
            padding: 1.25rem;
            text-align: left;
        }

        /* Form elements */
        form {
            display: flex;
            flex-direction: column;
            width: 100%%;
        }

        .form-group {
            display: flex;
            flex-direction: column;
            align-items: stretch;
            width: 100%%;
            margin-bottom: 1.125rem;
        }

        .form-group > * {
            width: 100%%;
        }

        .form-group:last-of-type {
            margin-bottom: 0;
        }

        .label-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 0.5rem;
            width: 100%%;
        }

        label {
            font-size: 0.8125rem;
            font-weight: 600;
            color: var(--text);
        }

        .badge {
            font-size: 0.5625rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            padding: 0.1875rem 0.5rem;
            border-radius: 4px;
            background: var(--line-green-light);
            color: var(--line-green);
        }

        input[type="text"],
        input[type="password"],
        select {
            display: block;
            width: 100%% !important;
            max-width: 100%%;
            min-width: 100%%;
            padding: 0.875rem 1rem;
            font-family: inherit;
            font-size: 0.9375rem;
            background: var(--bg-input);
            border: 1.5px solid var(--border);
            border-radius: 10px;
            color: var(--text);
            transition: all 0.2s ease;
            -webkit-appearance: none;
            -moz-appearance: none;
            appearance: none;
            box-sizing: border-box;
        }

        input[type="text"]::placeholder,
        input[type="password"]::placeholder {
            color: var(--text-muted);
        }

        input[type="text"]:focus,
        input[type="password"]:focus,
        select:focus {
            outline: none;
            background: var(--bg-card);
            border-color: var(--line-green);
            box-shadow: 0 0 0 3px rgba(6, 199, 85, 0.12);
        }

        input.mono {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.875rem;
            letter-spacing: -0.01em;
        }

        .input-hint {
            font-size: 0.75rem;
            color: var(--text-muted);
            margin-top: 0.5rem;
        }

        .input-hint a {
            color: var(--line-green);
            text-decoration: none;
        }

        .input-hint a:hover {
            text-decoration: underline;
        }

        /* Select dropdown */
        .select-wrapper {
            position: relative;
            display: flex;
            width: 100%%;
            align-items: center;
        }

        .select-wrapper select {
            flex: 1;
            width: 100%%;
            min-width: 0;
            padding-right: 2.75rem;
            cursor: pointer;
        }

        .select-wrapper::after {
            content: '';
            position: absolute;
            right: 1rem;
            top: 50%%;
            transform: translateY(-50%%);
            width: 0;
            height: 0;
            border-left: 5px solid transparent;
            border-right: 5px solid transparent;
            border-top: 6px solid var(--text-secondary);
            pointer-events: none;
            transition: border-color 0.2s ease;
        }

        .select-wrapper:hover::after {
            border-top-color: var(--text);
        }

        /* Buttons */
        .btn-group {
            display: flex;
            gap: 0.75rem;
            margin-top: 1.25rem;
        }

        button {
            flex: 1;
            padding: 0.75rem 1.25rem;
            font-family: inherit;
            font-size: 0.875rem;
            font-weight: 600;
            border-radius: 10px;
            cursor: pointer;
            transition: all 0.2s ease;
            border: none;
        }

        .btn-secondary {
            background: var(--bg-input);
            color: var(--text-secondary);
            border: 1px solid var(--border);
        }

        .btn-secondary:hover {
            background: var(--border);
            color: var(--text);
        }

        .btn-primary {
            background: var(--line-green);
            color: white;
            box-shadow: 0 4px 12px rgba(6, 199, 85, 0.25);
        }

        .btn-primary:hover {
            background: var(--line-green-dark);
            transform: translateY(-1px);
            box-shadow: 0 6px 16px rgba(6, 199, 85, 0.3);
        }

        .btn-primary:active {
            transform: translateY(0);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none !important;
        }

        /* Status toast */
        .status {
            position: fixed;
            bottom: 2rem;
            left: 50%%;
            transform: translateX(-50%%) translateY(20px);
            padding: 0.75rem 1.25rem;
            border-radius: 12px;
            font-size: 0.8125rem;
            font-weight: 500;
            align-items: center;
            gap: 0.5rem;
            opacity: 0;
            visibility: hidden;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            display: flex;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
            z-index: 100;
            white-space: nowrap;
        }

        .status.show {
            opacity: 1;
            visibility: visible;
            transform: translateX(-50%%) translateY(0);
        }

        .status.loading {
            background: var(--line-green-light);
            color: var(--line-green-dark);
        }

        .status.success {
            background: var(--success-light);
            color: var(--success);
        }

        .status.error {
            background: var(--error-light);
            color: var(--error);
        }

        .spinner {
            width: 14px;
            height: 14px;
            border: 2px solid currentColor;
            border-top-color: transparent;
            border-radius: 50%%;
            animation: spin 0.7s linear infinite;
        }

        @keyframes spin { to { transform: rotate(360deg); } }

        .status-icon { width: 14px; height: 14px; flex-shrink: 0; }

        /* Help section */
        .help-section {
            margin-top: 1.25rem;
            padding-top: 1rem;
            border-top: 1px solid var(--border);
        }

        .help-title {
            font-size: 0.6875rem;
            font-weight: 600;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.75rem;
        }

        .help-item {
            display: flex;
            align-items: flex-start;
            gap: 0.625rem;
            margin-bottom: 0.625rem;
            font-size: 0.8125rem;
            color: var(--text-secondary);
        }

        .help-item:last-child { margin-bottom: 0; }

        .help-icon {
            flex-shrink: 0;
            width: 20px;
            height: 20px;
            background: var(--bg-input);
            border-radius: 5px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.625rem;
            font-weight: 600;
            color: var(--text-muted);
        }

        .help-item a {
            color: var(--line-green);
            text-decoration: none;
        }

        .help-item a:hover {
            text-decoration: underline;
        }

        /* Footer */
        .github-link {
            margin-top: 1.5rem;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            text-decoration: none;
            color: var(--text-muted);
            font-size: 0.8125rem;
            font-weight: 500;
            transition: color 0.2s ease;
            width: 100%%;
        }

        .github-link:hover { color: var(--text-secondary); }
        .github-link svg { width: 16px; height: 16px; }

        @media (max-width: 480px) {
            .container { padding: 1.5rem 1rem; }
            .form-body { padding: 1rem; }
            .btn-group { flex-direction: column; }
            button { width: 100%%; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">
            <svg viewBox="0 0 55 55" xmlns="http://www.w3.org/2000/svg" style="height: 44px;">
                <path fill="#06C755" d="M42.6,55H12.4C5.6,55,0,49.4,0,42.6V12.4C0,5.6,5.6,0,12.4,0h30.2C49.4,0,55,5.6,55,12.4v30.2C55,49.4,49.4,55,42.6,55z"/>
                <path fill="#FFFFFF" d="M45.8,24.9c0-8.2-8.2-14.9-18.3-14.9C17.4,10,9.2,16.7,9.2,24.9c0,7.4,6.5,13.5,15.3,14.7c0.6,0.1,1.4,0.4,1.6,0.9c0.2,0.5,0.1,1.2,0.1,1.7c0,0-0.2,1.3-0.3,1.6c-0.1,0.5-0.4,1.8,1.6,1c2-0.8,10.6-6.2,14.4-10.6h0C44.6,31.2,45.8,28.2,45.8,24.9z"/>
                <path fill="#06C755" d="M39.7,29.6h-5.1h0c-0.2,0-0.4-0.2-0.4-0.4v0v0v-8v0v0c0-0.2,0.2-0.4,0.4-0.4h0h5.1c0.2,0,0.4,0.2,0.4,0.4v1.3c0,0.2-0.2,0.4-0.4,0.4h-3.5v1.4h3.5c0.2,0,0.4,0.2,0.4,0.4v1.3c0,0.2-0.2,0.4-0.4,0.4h-3.5v1.4h3.5c0.2,0,0.4,0.2,0.4,0.4v1.3C40.1,29.5,39.9,29.6,39.7,29.6z"/>
                <path fill="#06C755" d="M20.7,29.6c0.2,0,0.4-0.2,0.4-0.4V28c0-0.2-0.2-0.4-0.4-0.4h-3.5v-6.4c0-0.2-0.2-0.4-0.4-0.4h-1.3c-0.2,0-0.4,0.2-0.4,0.4v8v0v0c0,0.2,0.2,0.4,0.4,0.4h0H20.7z"/>
                <path fill="#06C755" d="M23.8,20.9h-1.3c-0.2,0-0.4,0.2-0.4,0.4v8c0,0.2,0.2,0.4,0.4,0.4h1.3c0.2,0,0.4-0.2,0.4-0.4v-8C24.1,21.1,24,20.9,23.8,20.9z"/>
                <path fill="#06C755" d="M32.6,20.9h-1.3c-0.2,0-0.4,0.2-0.4,0.4V26l-3.7-4.9c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0h-1.3c-0.2,0-0.4,0.2-0.4,0.4v8c0,0.2,0.2,0.4,0.4,0.4H27c0.2,0,0.4-0.2,0.4-0.4v-4.8l3.7,5c0,0,0.1,0.1,0.1,0.1c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0,0,0,0c0,0,0.1,0,0.1,0h1.3c0.2,0,0.4-0.2,0.4-0.4v-8C33,21.1,32.8,20.9,32.6,20.9z"/>
            </svg>
        </div>

        <div class="badge-wrapper">
            <div class="cli-badge">
                <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <rect x="2" y="2" width="12" height="12" rx="2" stroke="currentColor" stroke-width="1.5"/>
                    <path d="M5 6L7 8L5 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M9 10H11" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
                CLI Authentication
            </div>
        </div>

        <h1>Connect Your Account</h1>
        <p class="subtitle">Link your LINE Official Account to start using the CLI</p>

        <!-- Accounts section -->
        <div id="accountsSection" class="accounts-section hidden">
            <div class="section-header">
                <span class="section-title">Connected Accounts</span>
                <span id="accountCount" class="account-count">0 accounts</span>
            </div>
            <div id="accountsList" class="accounts-list"></div>
            <button id="addAccountBtn" class="add-account-btn">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                Add Account
            </button>
        </div>

        <!-- Empty state -->
        <div id="emptyState" class="empty-state hidden">
            <div class="empty-state-icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            </div>
            <h3>No accounts connected</h3>
            <p>Add your first LINE channel to get started</p>
        </div>

        <!-- Setup form card -->
        <div id="setupCard" class="form-card hidden">
            <div class="form-header">
                <h2>Add LINE Account</h2>
                <button id="closeSetupBtn" class="close-btn">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
            </div>
            <div class="form-body">
                <form id="setupForm" autocomplete="off">
                    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">

                    <div class="form-group">
                        <div class="label-row">
                            <label for="apiType">API Type</label>
                            <span class="badge">Required</span>
                        </div>
                        <div class="select-wrapper">
                            <select id="apiType" name="api_type" required>
                                <option value="messaging">Messaging API</option>
                                <option value="liff">LIFF</option>
                                <option value="login">LINE Login</option>
                            </select>
                        </div>
                        <div class="input-hint">Select the type of LINE API you want to use</div>
                    </div>

                    <div class="form-group">
                        <div class="label-row">
                            <label for="accountName">Account Name</label>
                        </div>
                        <input
                            type="text"
                            id="accountName"
                            name="account_name"
                            placeholder="e.g., my-shop, production"
                            value="default"
                        >
                        <div class="input-hint">A friendly name to identify this channel</div>
                    </div>

                    <div class="form-group">
                        <div class="label-row">
                            <label for="accessToken">Channel Access Token</label>
                            <span class="badge">Required</span>
                        </div>
                        <input
                            type="password"
                            id="accessToken"
                            name="access_token"
                            class="mono"
                            placeholder="Paste your long-lived channel access token"
                            required
                        >
                        <div class="input-hint">
                            Found in <a href="https://developers.line.biz/console/" target="_blank" rel="noopener noreferrer">LINE Developers Console</a> &rarr; Messaging API tab
                        </div>
                    </div>

                    <div class="btn-group">
                        <button type="button" id="testBtn" class="btn-secondary">Test Connection</button>
                        <button type="submit" id="submitBtn" class="btn-primary">Save & Connect</button>
                    </div>

                    <div id="status" class="status"></div>
                </form>

                <div class="help-section">
                    <div class="help-title">Where to find your token</div>
                    <div class="help-item">
                        <span class="help-icon">1</span>
                        <span>Go to <a href="https://developers.line.biz/console/" target="_blank" rel="noopener noreferrer">LINE Developers Console</a></span>
                    </div>
                    <div class="help-item">
                        <span class="help-icon">2</span>
                        <span>Select your Provider and Messaging API Channel</span>
                    </div>
                    <div class="help-item">
                        <span class="help-icon">3</span>
                        <span>In the Messaging API tab, click "Issue" for Channel access token</span>
                    </div>
                </div>
            </div>
        </div>

        <a href="https://github.com/salmonumbrella/line-official-cli" target="_blank" rel="noopener noreferrer" class="github-link">
            <svg viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            View on GitHub
        </a>
    </div>

    <script>
        const csrfToken = '{{.CSRFToken}}';
        const accountsSection = document.getElementById('accountsSection');
        const accountsList = document.getElementById('accountsList');
        const accountCount = document.getElementById('accountCount');
        const emptyState = document.getElementById('emptyState');
        const addAccountBtn = document.getElementById('addAccountBtn');
        const setupCard = document.getElementById('setupCard');
        const closeSetupBtn = document.getElementById('closeSetupBtn');
        const form = document.getElementById('setupForm');
        const testBtn = document.getElementById('testBtn');
        const submitBtn = document.getElementById('submitBtn');
        const status = document.getElementById('status');

        let accounts = [];

        async function loadAccounts() {
            try {
                const response = await fetch('/accounts');
                const data = await response.json();
                accounts = data.accounts || [];
                renderAccounts();
            } catch (err) {
                accounts = [];
                renderAccounts();
            }
        }

        function renderAccounts() {
            accountCount.textContent = accounts.length + ' account' + (accounts.length !== 1 ? 's' : '');
            if (accounts.length > 0) {
                closeSetupBtn.classList.add('show');
            } else {
                closeSetupBtn.classList.remove('show');
            }
            if (accounts.length === 0) {
                accountsSection.classList.add('hidden');
                emptyState.classList.remove('hidden');
                setupCard.classList.remove('hidden');
            } else {
                accountsSection.classList.remove('hidden');
                emptyState.classList.add('hidden');
                setupCard.classList.add('hidden');
                accountsList.innerHTML = '';
                accounts.forEach(function(acc) {
                    var card = document.createElement('div');
                    card.className = 'account-card' + (acc.isPrimary ? ' primary' : '');
                    card.dataset.name = acc.name;

                    var avatar = document.createElement('div');
                    avatar.className = 'account-avatar';
                    avatar.textContent = acc.name.charAt(0).toUpperCase();

                    var info = document.createElement('div');
                    info.className = 'account-info';

                    var nameDiv = document.createElement('div');
                    nameDiv.className = 'account-name';
                    nameDiv.textContent = acc.name;

                    var botDiv = document.createElement('div');
                    botDiv.className = 'account-bot';
                    botDiv.textContent = acc.botName || '';

                    info.appendChild(nameDiv);
                    info.appendChild(botDiv);

                    var actions = document.createElement('div');
                    actions.className = 'account-actions';

                    if (acc.isPrimary) {
                        var badge = document.createElement('span');
                        badge.className = 'primary-badge';
                        badge.innerHTML = '<svg viewBox="0 0 24 24" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>Primary';
                        actions.appendChild(badge);
                    } else {
                        var setPrimaryBtn = document.createElement('button');
                        setPrimaryBtn.className = 'set-primary-btn';
                        setPrimaryBtn.textContent = 'Set as primary';
                        setPrimaryBtn.addEventListener('click', function() { setPrimary(acc.name); });
                        actions.appendChild(setPrimaryBtn);
                    }

                    var removeBtn = document.createElement('button');
                    removeBtn.className = 'remove-btn';
                    removeBtn.title = 'Remove account';
                    removeBtn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>';
                    removeBtn.addEventListener('click', function() { removeAccount(acc.name); });
                    actions.appendChild(removeBtn);

                    card.appendChild(avatar);
                    card.appendChild(info);
                    card.appendChild(actions);
                    accountsList.appendChild(card);
                });
            }
        }

        async function setPrimary(name) {
            try {
                const response = await fetch('/set-primary', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({ name })
                });
                const data = await response.json();
                if (data.success) await loadAccounts();
            } catch (err) {
                console.error('Failed to set primary:', err);
            }
        }

        async function removeAccount(name) {
            if (!confirm('Remove "' + name + '" from LINE CLI?')) return;
            try {
                const response = await fetch('/remove-account', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
                    body: JSON.stringify({ name })
                });
                const data = await response.json();
                if (data.success) await loadAccounts();
            } catch (err) {
                console.error('Failed to remove account:', err);
            }
        }

        addAccountBtn.addEventListener('click', function() {
            setupCard.classList.remove('hidden');
            document.getElementById('accountName').focus();
        });

        closeSetupBtn.addEventListener('click', function() {
            if (accounts.length > 0) {
                setupCard.classList.add('hidden');
                form.reset();
                hideStatus();
            }
        });

        function showStatus(type, message) {
            status.className = 'status show ' + type;
            if (type === 'loading') {
                status.innerHTML = '<div class="spinner"></div><span>' + message + '</span>';
            } else {
                var icon = type === 'success'
                    ? '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M13 5L6.5 11.5L3 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>'
                    : '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M12 4L4 12M4 4L12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>';
                status.innerHTML = icon + '<span>' + message + '</span>';
            }
        }

        function hideStatus() {
            status.className = 'status';
        }

        function getFormData() {
            return {
                account_name: document.getElementById('accountName').value.trim() || 'default',
                access_token: document.getElementById('accessToken').value.trim(),
                api_type: document.getElementById('apiType').value
            };
        }

        testBtn.addEventListener('click', async function() {
            var data = getFormData();

            if (!data.access_token) {
                showStatus('error', 'Please enter your channel access token');
                return;
            }

            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Testing connection...');

            try {
                var response = await fetch('/validate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify(data)
                });

                var result = await response.json();

                if (result.success) {
                    showStatus('success', 'Connected! Bot: ' + result.bot_name);
                } else {
                    showStatus('error', result.error);
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
            } finally {
                testBtn.disabled = false;
                submitBtn.disabled = false;
            }
        });

        form.addEventListener('submit', async function(e) {
            e.preventDefault();

            var data = getFormData();

            if (!data.access_token) {
                showStatus('error', 'Please enter your channel access token');
                return;
            }

            submitBtn.disabled = true;
            testBtn.disabled = true;
            showStatus('loading', 'Saving credentials...');

            try {
                var response = await fetch('/submit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify(data)
                });

                var result = await response.json();

                if (result.success) {
                    showStatus('success', 'Credentials saved! Redirecting...');
                    setTimeout(function() {
                        window.location.href = '/success?name=' + encodeURIComponent(result.account_name) + '&bot=' + encodeURIComponent(result.bot_name || '');
                    }, 1000);
                } else {
                    showStatus('error', result.error);
                    submitBtn.disabled = false;
                    testBtn.disabled = false;
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
                submitBtn.disabled = false;
                testBtn.disabled = false;
            }
        });

        // Load accounts on page load
        loadAccounts();
    </script>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Setup Complete - LINE CLI</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #0F0F0F;
            --bg-card: #1A1A1A;
            --bg-terminal: #0A0A0A;
            --border: #333333;
            --text: #FFFFFF;
            --text-secondary: #B3B3B3;
            --text-muted: #737373;
            --line-green: #06C755;
            --line-green-dark: #05A847;
            --line-green-light: rgba(6, 199, 85, 0.15);
            --success: #06C755;
            --success-light: rgba(6, 199, 85, 0.15);
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        html { height: 100%%; }

        body {
            font-family: 'Plus Jakarta Sans', -apple-system, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%%;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem 3rem;
            position: relative;
        }

        body::before {
            content: '';
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            background:
                radial-gradient(ellipse at top, rgba(6, 199, 85, 0.08) 0%%, transparent 50%%),
                radial-gradient(ellipse at bottom right, rgba(6, 199, 85, 0.04) 0%%, transparent 40%%);
            pointer-events: none;
        }

        .container {
            width: 100%%;
            max-width: 420px;
            text-align: center;
            position: relative;
            z-index: 1;
        }

        .success-icon {
            width: 72px;
            height: 72px;
            background: linear-gradient(135deg, var(--success-light) 0%%, #BBF7D0 100%%);
            border-radius: 50%%;
            margin: 0 auto 1.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: scaleIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
            box-shadow: 0 8px 24px rgba(6, 199, 85, 0.2);
        }

        @keyframes scaleIn {
            from { transform: scale(0); opacity: 0; }
            to { transform: scale(1); opacity: 1; }
        }

        .success-icon svg {
            width: 36px;
            height: 36px;
            color: var(--success);
        }

        h1 {
            font-size: 1.5rem;
            font-weight: 700;
            margin-bottom: 0.25rem;
            letter-spacing: -0.02em;
            animation: fadeUp 0.5s ease 0.15s both;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.9375rem;
            margin-bottom: 1rem;
            animation: fadeUp 0.5s ease 0.2s both;
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .account-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            background: var(--line-green-light);
            color: var(--line-green);
            font-size: 0.875rem;
            font-weight: 600;
            padding: 0.5rem 1rem;
            border-radius: 100px;
            margin-bottom: 1.25rem;
            animation: fadeUp 0.5s ease 0.25s both;
        }

        .account-badge .dot {
            width: 8px;
            height: 8px;
            background: var(--success);
            border-radius: 50%%;
            animation: dotPulse 2s ease-in-out infinite;
        }

        @keyframes dotPulse {
            0%%, 100%% { opacity: 1; transform: scale(1); }
            50%% { opacity: 0.6; transform: scale(0.9); }
        }

        .terminal {
            background: var(--bg-terminal);
            border-radius: 12px;
            overflow: hidden;
            text-align: left;
            animation: fadeUp 0.5s ease 0.3s both;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
        }

        .terminal-bar {
            background: #1F2937;
            padding: 0.75rem 1rem;
            display: flex;
            align-items: center;
            gap: 0.375rem;
        }

        .terminal-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%%;
        }

        .terminal-dot.red { background: #EF4444; }
        .terminal-dot.yellow { background: #F59E0B; }
        .terminal-dot.green { background: #10B981; }

        .terminal-body { padding: 1rem; }

        .terminal-line {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.8125rem;
            margin-bottom: 0.5rem;
            color: #E5E7EB;
        }

        .terminal-line:last-child { margin-bottom: 0; }
        .terminal-prompt { color: var(--line-green); user-select: none; }
        .terminal-cmd { color: #60A5FA; }
        .terminal-output {
            color: #9CA3AF;
            padding-left: 1rem;
            margin-top: -0.25rem;
            margin-bottom: 0.625rem;
            font-size: 0.75rem;
        }

        .terminal-cursor {
            display: inline-block;
            width: 8px;
            height: 16px;
            background: var(--line-green);
            animation: cursorBlink 1.2s step-end infinite;
            margin-left: 2px;
            vertical-align: middle;
        }

        @keyframes cursorBlink {
            0%%, 50%% { opacity: 1; }
            50.01%%, 100%% { opacity: 0; }
        }

        .message {
            margin-top: 1.25rem;
            padding: 1rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            text-align: center;
            animation: fadeUp 0.5s ease 0.4s both;
        }

        .message-icon {
            font-size: 1.25rem;
            margin-bottom: 0.25rem;
        }

        .message-title {
            font-weight: 600;
            font-size: 0.9375rem;
            margin-bottom: 0.125rem;
            color: var(--text);
        }

        .message-text {
            font-size: 0.8125rem;
            color: var(--text-secondary);
            line-height: 1.5;
        }

        .message-text code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--line-green-light);
            color: var(--line-green);
            padding: 0.125rem 0.375rem;
            border-radius: 4px;
            font-size: 0.75rem;
        }

        .github-link {
            margin-top: 1.5rem;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            text-decoration: none;
            color: var(--text-muted);
            font-size: 0.8125rem;
            font-weight: 500;
            transition: color 0.2s ease;
            width: 100%%;
        }

        .github-link:hover { color: var(--text-secondary); }
        .github-link svg { width: 16px; height: 16px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 32 32" fill="none">
                <path d="M8 16L14 22L24 10" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </div>

        <h1>You're all set!</h1>
        <p class="subtitle">LINE CLI is connected and ready to use</p>

        <div class="account-badge">
            <span class="dot"></span>
            <span>{{.AccountName}}</span>
        </div>

        {{if .BotName}}
        <p class="subtitle" style="margin-top: -0.5rem; font-size: 0.875rem;">Bot: {{.BotName}}</p>
        {{end}}

        <div class="terminal">
            <div class="terminal-bar">
                <span class="terminal-dot red"></span>
                <span class="terminal-dot yellow"></span>
                <span class="terminal-dot green"></span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">line</span>
                    <span>message quota</span>
                </div>
                <div class="terminal-output">Checking message quota...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">line</span>
                    <span>richmenu list</span>
                </div>
                <div class="terminal-output">Listing rich menus...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cursor"></span>
                </div>
            </div>
        </div>

        <div class="message">
            <div class="message-icon">&larr;</div>
            <div class="message-title">Return to your terminal</div>
            <div class="message-text">You can close this window and start using the CLI.<br>Try running <code>line --help</code> to see all commands.</div>
        </div>

        <a href="https://github.com/salmonumbrella/line-official-cli" target="_blank" rel="noopener noreferrer" class="github-link">
            <svg viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
            </svg>
            View on GitHub
        </a>
    </div>

    <script>
        // Signal completion to server
        window.addEventListener('load', function() {
            setTimeout(function() {
                fetch('/complete', { method: 'POST' }).catch(function() {});
            }, 500);
        });
    </script>
</body>
</html>`
