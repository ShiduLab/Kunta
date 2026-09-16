package io.github.shidulab.kunta;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.provider.OpenableColumns;
import android.webkit.JavascriptInterface;
import android.webkit.ValueCallback;
import android.webkit.WebChromeClient;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.HashSet;
import java.util.Set;
import java.util.zip.GZIPInputStream;

public class MainActivity extends Activity {
    private static final int FILE_REQUEST = 7001;
    private static final String APP_URL = "file:///android_asset/index.html";

    private WebView webView;
    private ValueCallback<Uri[]> fileCallback;
    private String pendingSharedText;
    private boolean nativeTextPicker;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        getWindow().setStatusBarColor(Color.rgb(16, 20, 15));
        getWindow().setNavigationBarColor(Color.rgb(16, 20, 15));
        pendingSharedText = extractSharedText(getIntent());

        webView = new WebView(this);
        webView.setBackgroundColor(Color.rgb(16, 20, 15));
        setContentView(webView);

        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setAllowFileAccess(true);
        settings.setAllowContentAccess(true);
        settings.setBuiltInZoomControls(false);
        settings.setDisplayZoomControls(false);
        settings.setUseWideViewPort(true);
        settings.setTextZoom(100);

        webView.addJavascriptInterface(new NativeBridge(), "KuntaNative");

        webView.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onShowFileChooser(WebView view, ValueCallback<Uri[]> callback, FileChooserParams params) {
                if (fileCallback != null) fileCallback.onReceiveValue(null);
                fileCallback = callback;
                nativeTextPicker = false;
                launchTextPicker();
                return true;
            }
        });

        webView.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest request) {
                return handleUri(request.getUrl());
            }

            @Override
            public boolean shouldOverrideUrlLoading(WebView view, String url) {
                return handleUri(Uri.parse(url));
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                super.onPageFinished(view, url);
                deliverSharedText();
            }
        });

        if (savedInstanceState == null) webView.loadUrl(APP_URL);
        else webView.restoreState(savedInstanceState);
    }

    private void launchTextPicker() {
        Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        intent.setType("*/*");
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        startActivityForResult(intent, FILE_REQUEST);
    }

    private String displayName(Uri uri) {
        if (uri == null) return "file";
        try (android.database.Cursor cursor = getContentResolver().query(uri, null, null, null, null)) {
            if (cursor != null && cursor.moveToFirst()) {
                int idx = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME);
                if (idx >= 0) {
                    String name = cursor.getString(idx);
                    if (name != null && !name.isEmpty()) return name;
                }
            }
        } catch (Exception ignored) {}
        String tail = uri.getLastPathSegment();
        return tail == null || tail.isEmpty() ? "file" : tail;
    }

    private String readTextUri(Uri uri) throws Exception {
        StringBuilder out = new StringBuilder();
        try (InputStream raw = getContentResolver().openInputStream(uri)) {
            if (raw == null) throw new IllegalStateException("Impossibile aprire il file selezionato.");
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(raw, StandardCharsets.UTF_8), 64 * 1024)) {
                char[] buffer = new char[16 * 1024];
                int n;
                while ((n = reader.read(buffer)) != -1) out.append(buffer, 0, n);
            }
        }
        if (out.length() > 0 && out.charAt(0) == '\uFEFF') out.deleteCharAt(0);
        return out.toString();
    }

    private void deliverTextToWeb(String text, String status) {
        if (webView == null) return;
        String quotedText = JSONObject.quote(text == null ? "" : text);
        String quotedStatus = JSONObject.quote(status == null ? "" : status);
        webView.evaluateJavascript(
                "(function(){" +
                        "var e=document.getElementById('inputText');" +
                        "if(e){e.value=" + quotedText + ";e.dispatchEvent(new Event('input',{bubbles:true}));}" +
                        "var s=document.getElementById('status');" +
                        "if(s)s.textContent=" + quotedStatus + ";" +
                        "})();",
                null);
    }

    private boolean handleUri(Uri uri) {
        if (uri == null) return false;
        String url = uri.toString();
        if (url.startsWith("file:///android_asset/")) return false;
        String scheme = uri.getScheme();
        if ("http".equalsIgnoreCase(scheme) || "https".equalsIgnoreCase(scheme)) {
            startActivity(new Intent(Intent.ACTION_VIEW, uri));
            return true;
        }
        return false;
    }

    private String extractSharedText(Intent intent) {
        if (intent == null || !Intent.ACTION_SEND.equals(intent.getAction())) return null;
        CharSequence text = intent.getCharSequenceExtra(Intent.EXTRA_TEXT);
        return text == null ? null : text.toString();
    }

    private void deliverSharedText() {
        if (pendingSharedText == null || pendingSharedText.isEmpty() || webView == null) return;
        String text = pendingSharedText;
        pendingSharedText = null;
        deliverTextToWeb(text, "Testo ricevuto dalla condivisione.");
    }

    @Override
    protected void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        setIntent(intent);
        pendingSharedText = extractSharedText(intent);
        deliverSharedText();
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode != FILE_REQUEST) return;

        Uri uri = resultCode == RESULT_OK && data != null ? data.getData() : null;

        if (nativeTextPicker) {
            nativeTextPicker = false;
            if (uri == null) return;
            final Uri selected = uri;
            new Thread(() -> {
                try {
                    String text = readTextUri(selected);
                    String name = displayName(selected);
                    runOnUiThread(() -> deliverTextToWeb(text, "Aperto: " + name));
                } catch (Exception e) {
                    String error = "Errore apertura file: " + e.getMessage();
                    runOnUiThread(() -> deliverTextToWeb("", error));
                }
            }, "KuntaFileReader").start();
            return;
        }

        if (fileCallback != null) {
            Uri[] result = uri == null ? null : new Uri[]{uri};
            fileCallback.onReceiveValue(result);
            fileCallback = null;
        }
    }

    @Override
    protected void onSaveInstanceState(Bundle outState) {
        if (webView != null) webView.saveState(outState);
        super.onSaveInstanceState(outState);
    }

    @Override
    @SuppressWarnings("deprecation")
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) webView.goBack();
        else super.onBackPressed();
    }

    private class NativeBridge {
        @JavascriptInterface
        public void openTextFile() {
            runOnUiThread(() -> {
                if (fileCallback != null) {
                    fileCallback.onReceiveValue(null);
                    fileCallback = null;
                }
                nativeTextPicker = true;
                launchTextPicker();
            });
        }

        @JavascriptInterface
        public String readClipboard() {
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard == null || !clipboard.hasPrimaryClip() || clipboard.getPrimaryClip() == null || clipboard.getPrimaryClip().getItemCount() == 0) return "";
            CharSequence text = clipboard.getPrimaryClip().getItemAt(0).coerceToText(MainActivity.this);
            return text == null ? "" : text.toString();
        }

        @JavascriptInterface
        public void writeClipboard(String text) {
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            if (clipboard != null) clipboard.setPrimaryClip(ClipData.newPlainText("Kunta", text == null ? "" : text));
        }

        @JavascriptInterface
        public String lookupAnagrams(String signaturesJson) {
            JSONObject result = new JSONObject();
            try {
                JSONArray requested = new JSONArray(signaturesJson == null ? "[]" : signaturesJson);
                Set<String> wanted = new HashSet<>();
                for (int i = 0; i < requested.length(); i++) {
                    String sig = requested.optString(i, "");
                    if (!sig.isEmpty()) wanted.add(sig);
                }
                if (wanted.isEmpty()) return result.toString();

                try (InputStream raw = getAssets().open("assets/anagram_index.tsv.gz");
                     GZIPInputStream gz = new GZIPInputStream(raw);
                     BufferedReader reader = new BufferedReader(new InputStreamReader(gz, StandardCharsets.UTF_8), 64 * 1024)) {
                    String line;
                    while ((line = reader.readLine()) != null && !wanted.isEmpty()) {
                        int tab = line.indexOf('\t');
                        if (tab <= 0) continue;
                        String sig = line.substring(0, tab);
                        if (!wanted.contains(sig)) continue;
                        JSONArray words = new JSONArray();
                        String payload = line.substring(tab + 1);
                        if (!payload.isEmpty()) {
                            for (String word : payload.split("\\|")) words.put(word);
                        }
                        result.put(sig, words);
                        wanted.remove(sig);
                    }
                }
            } catch (Exception e) {
                try { result.put("__error", e.getClass().getSimpleName() + ": " + e.getMessage()); } catch (Exception ignored) {}
            }
            return result.toString();
        }

        @JavascriptInterface
        public void share(String text) {
            runOnUiThread(() -> {
                Intent send = new Intent(Intent.ACTION_SEND);
                send.setType("text/plain");
                send.putExtra(Intent.EXTRA_TEXT, text == null ? "" : text);
                startActivity(Intent.createChooser(send, "Condividi risultato"));
            });
        }
    }
}
