package io.github.shidulab.kunta;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.content.res.AssetFileDescriptor;
import android.graphics.Color;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.provider.OpenableColumns;
import android.text.Html;
import android.util.Xml;
import android.webkit.JavascriptInterface;
import android.webkit.ValueCallback;
import android.webkit.WebChromeClient;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import com.tom_roush.pdfbox.android.PDFBoxResourceLoader;
import com.tom_roush.pdfbox.pdmodel.PDDocument;
import com.tom_roush.pdfbox.text.PDFTextStripper;

import org.json.JSONArray;
import org.json.JSONObject;
import org.xmlpull.v1.XmlPullParser;

import java.io.BufferedReader;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.FileInputStream;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;
import java.util.List;
import java.util.Locale;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

public class MainActivity extends Activity {
    private static final int FILE_REQUEST = 7001;
    private static final String APP_URL = "file:///android_asset/index.html";
    private static final int MAX_DOCUMENT_BYTES = 96 * 1024 * 1024;

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

        PDFBoxResourceLoader.init(getApplicationContext());

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

    private String extensionOf(String name) {
        if (name == null) return "";
        int dot = name.lastIndexOf('.');
        if (dot < 0 || dot == name.length() - 1) return "";
        return name.substring(dot + 1).toLowerCase(Locale.ROOT);
    }

    private byte[] readAllBytes(Uri uri) throws Exception {
        try (InputStream in = getContentResolver().openInputStream(uri)) {
            if (in == null) throw new IllegalStateException("Impossibile aprire il file selezionato.");
            ByteArrayOutputStream out = new ByteArrayOutputStream();
            byte[] buffer = new byte[32 * 1024];
            int total = 0;
            int n;
            while ((n = in.read(buffer)) != -1) {
                total += n;
                if (total > MAX_DOCUMENT_BYTES) {
                    throw new IllegalStateException("Documento troppo grande per l'apertura diretta (limite 96 MB).");
                }
                out.write(buffer, 0, n);
            }
            return out.toByteArray();
        }
    }

    private String readTextUri(Uri uri) throws Exception {
        String name = displayName(uri);
        String ext = extensionOf(name);
        String mime = getContentResolver().getType(uri);
        if (mime == null) mime = "";

        if ("pdf".equals(ext) || "application/pdf".equalsIgnoreCase(mime)) {
            return extractPdf(uri);
        }

        if ("docx".equals(ext)) return extractDocx(uri);
        if ("odt".equals(ext) || "ods".equals(ext) || "odp".equals(ext)) return extractOpenDocument(uri);
        if ("pptx".equals(ext)) return extractPptx(uri);
        if ("xlsx".equals(ext)) return extractXlsx(uri);
        if ("epub".equals(ext)) return extractEpub(uri);

        if ("doc".equals(ext)) {
            String converted = tryProviderTextConversion(uri);
            if (converted != null && !converted.trim().isEmpty()) return converted;
            return extractLegacyDoc(readAllBytes(uri));
        }

        byte[] bytes = readAllBytes(uri);
        String decoded = decodeText(bytes);

        if ("rtf".equals(ext) || mime.toLowerCase(Locale.ROOT).contains("rtf")) {
            return stripRtf(decoded);
        }
        if ("html".equals(ext) || "htm".equals(ext) || mime.toLowerCase(Locale.ROOT).contains("html")) {
            return htmlToText(decoded);
        }
        if ("xml".equals(ext) || mime.toLowerCase(Locale.ROOT).contains("xml")) {
            String xmlText = xmlToText(bytes);
            return xmlText.trim().isEmpty() ? decoded : xmlText;
        }

        if (looksBinary(bytes)) {
            String converted = tryProviderTextConversion(uri);
            if (converted != null && !converted.trim().isEmpty()) return converted;
            throw new IllegalStateException("Il file contiene dati binari non riconosciuti come documento testuale.");
        }

        return decoded;
    }

    private String tryProviderTextConversion(Uri uri) {
        try (AssetFileDescriptor afd = getContentResolver().openTypedAssetFileDescriptor(uri, "text/plain", null)) {
            if (afd == null) return null;
            try (FileInputStream in = afd.createInputStream()) {
                ByteArrayOutputStream out = new ByteArrayOutputStream();
                byte[] buffer = new byte[16 * 1024];
                int n;
                int total = 0;
                while ((n = in.read(buffer)) != -1) {
                    total += n;
                    if (total > MAX_DOCUMENT_BYTES) break;
                    out.write(buffer, 0, n);
                }
                return decodeText(out.toByteArray());
            }
        } catch (Exception ignored) {
            return null;
        }
    }

    private String extractPdf(Uri uri) throws Exception {
        try (InputStream in = getContentResolver().openInputStream(uri)) {
            if (in == null) throw new IllegalStateException("Impossibile aprire il PDF.");
            try (PDDocument document = PDDocument.load(in)) {
                String text = new PDFTextStripper().getText(document);
                if (text == null || text.trim().isEmpty()) {
                    throw new IllegalStateException("PDF senza testo estraibile. Se è una scansione serve OCR.");
                }
                return normalizeExtractedText(text);
            }
        }
    }

    private String extractDocx(Uri uri) throws Exception {
        List<ZipTextPart> parts = new ArrayList<>();
        try (InputStream in = getContentResolver().openInputStream(uri);
             ZipInputStream zip = new ZipInputStream(in)) {
            ZipEntry entry;
            while ((entry = zip.getNextEntry()) != null) {
                String n = entry.getName();
                if (n.equals("word/document.xml") ||
                        n.matches("word/header\\d+\\.xml") ||
                        n.matches("word/footer\\d+\\.xml") ||
                        n.equals("word/footnotes.xml") || n.equals("word/endnotes.xml") || n.equals("word/comments.xml")) {
                    parts.add(new ZipTextPart(n, readCurrentZipEntry(zip)));
                }
                zip.closeEntry();
            }
        }
        Collections.sort(parts);
        StringBuilder out = new StringBuilder();
        for (ZipTextPart p : parts) appendBlock(out, xmlToText(p.bytes));
        if (out.length() == 0) throw new IllegalStateException("DOCX senza testo estraibile.");
        return normalizeExtractedText(out.toString());
    }

    private String extractOpenDocument(Uri uri) throws Exception {
        byte[] content = findZipEntry(uri, "content.xml");
        if (content == null) throw new IllegalStateException("Documento OpenDocument privo di content.xml.");
        String text = xmlToText(content);
        if (text.trim().isEmpty()) throw new IllegalStateException("Documento senza testo estraibile.");
        return normalizeExtractedText(text);
    }

    private String extractPptx(Uri uri) throws Exception {
        List<ZipTextPart> slides = new ArrayList<>();
        try (InputStream in = getContentResolver().openInputStream(uri);
             ZipInputStream zip = new ZipInputStream(in)) {
            ZipEntry entry;
            while ((entry = zip.getNextEntry()) != null) {
                String n = entry.getName();
                if (n.matches("ppt/slides/slide\\d+\\.xml") || n.matches("ppt/notesSlides/notesSlide\\d+\\.xml")) {
                    slides.add(new ZipTextPart(n, readCurrentZipEntry(zip)));
                }
                zip.closeEntry();
            }
        }
        Collections.sort(slides);
        StringBuilder out = new StringBuilder();
        for (ZipTextPart slide : slides) appendBlock(out, xmlToText(slide.bytes));
        if (out.length() == 0) throw new IllegalStateException("PPTX senza testo estraibile.");
        return normalizeExtractedText(out.toString());
    }

    private String extractXlsx(Uri uri) throws Exception {
        StringBuilder out = new StringBuilder();
        try (InputStream in = getContentResolver().openInputStream(uri);
             ZipInputStream zip = new ZipInputStream(in)) {
            ZipEntry entry;
            while ((entry = zip.getNextEntry()) != null) {
                String n = entry.getName();
                if (n.equals("xl/sharedStrings.xml") || n.matches("xl/worksheets/sheet\\d+\\.xml")) {
                    appendBlock(out, xmlToText(readCurrentZipEntry(zip)));
                }
                zip.closeEntry();
            }
        }
        if (out.length() == 0) throw new IllegalStateException("XLSX senza testo estraibile.");
        return normalizeExtractedText(out.toString());
    }

    private String extractEpub(Uri uri) throws Exception {
        List<ZipTextPart> pages = new ArrayList<>();
        try (InputStream in = getContentResolver().openInputStream(uri);
             ZipInputStream zip = new ZipInputStream(in)) {
            ZipEntry entry;
            while ((entry = zip.getNextEntry()) != null) {
                String n = entry.getName().toLowerCase(Locale.ROOT);
                if (n.endsWith(".xhtml") || n.endsWith(".html") || n.endsWith(".htm")) {
                    pages.add(new ZipTextPart(entry.getName(), readCurrentZipEntry(zip)));
                }
                zip.closeEntry();
            }
        }
        Collections.sort(pages);
        StringBuilder out = new StringBuilder();
        for (ZipTextPart page : pages) appendBlock(out, htmlToText(decodeText(page.bytes)));
        if (out.length() == 0) throw new IllegalStateException("EPUB senza testo estraibile.");
        return normalizeExtractedText(out.toString());
    }

    private byte[] findZipEntry(Uri uri, String wanted) throws Exception {
        try (InputStream in = getContentResolver().openInputStream(uri);
             ZipInputStream zip = new ZipInputStream(in)) {
            ZipEntry entry;
            while ((entry = zip.getNextEntry()) != null) {
                if (wanted.equals(entry.getName())) return readCurrentZipEntry(zip);
                zip.closeEntry();
            }
        }
        return null;
    }

    private byte[] readCurrentZipEntry(ZipInputStream zip) throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        byte[] buffer = new byte[16 * 1024];
        int n;
        while ((n = zip.read(buffer)) != -1) out.write(buffer, 0, n);
        return out.toByteArray();
    }

    private String xmlToText(byte[] xmlBytes) {
        StringBuilder out = new StringBuilder();
        try {
            XmlPullParser parser = Xml.newPullParser();
            parser.setInput(new ByteArrayInputStream(xmlBytes), null);
            int event = parser.getEventType();
            while (event != XmlPullParser.END_DOCUMENT) {
                if (event == XmlPullParser.TEXT) {
                    String text = parser.getText();
                    if (text != null && !text.isEmpty()) out.append(text);
                } else if (event == XmlPullParser.END_TAG) {
                    String tag = parser.getName();
                    if ("p".equalsIgnoreCase(tag) || "tr".equalsIgnoreCase(tag) ||
                            "paragraph".equalsIgnoreCase(tag) || "h".equalsIgnoreCase(tag)) {
                        out.append('\n');
                    } else if ("tc".equalsIgnoreCase(tag) || "tab".equalsIgnoreCase(tag)) {
                        out.append('\t');
                    }
                }
                event = parser.next();
            }
        } catch (Exception ignored) {
            return "";
        }
        return normalizeExtractedText(out.toString());
    }

    private String htmlToText(String html) {
        if (html == null) return "";
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            return normalizeExtractedText(Html.fromHtml(html, Html.FROM_HTML_MODE_LEGACY).toString());
        }
        return normalizeExtractedText(Html.fromHtml(html).toString());
    }

    private String decodeText(byte[] bytes) {
        if (bytes == null || bytes.length == 0) return "";
        if (bytes.length >= 3 && (bytes[0] & 0xff) == 0xef && (bytes[1] & 0xff) == 0xbb && (bytes[2] & 0xff) == 0xbf) {
            return new String(bytes, 3, bytes.length - 3, StandardCharsets.UTF_8);
        }
        if (bytes.length >= 2 && (bytes[0] & 0xff) == 0xff && (bytes[1] & 0xff) == 0xfe) {
            return new String(bytes, 2, bytes.length - 2, StandardCharsets.UTF_16LE);
        }
        if (bytes.length >= 2 && (bytes[0] & 0xff) == 0xfe && (bytes[1] & 0xff) == 0xff) {
            return new String(bytes, 2, bytes.length - 2, StandardCharsets.UTF_16BE);
        }

        String utf8 = new String(bytes, StandardCharsets.UTF_8);
        int replacements = 0;
        for (int i = 0; i < utf8.length(); i++) if (utf8.charAt(i) == '\uFFFD') replacements++;
        if (replacements <= Math.max(1, utf8.length() / 500)) return utf8;
        return new String(bytes, Charset.forName("windows-1252"));
    }

    private boolean looksBinary(byte[] bytes) {
        if (bytes == null || bytes.length == 0) return false;
        int sample = Math.min(bytes.length, 8192);
        int controls = 0;
        int nul = 0;
        for (int i = 0; i < sample; i++) {
            int b = bytes[i] & 0xff;
            if (b == 0) nul++;
            if (b < 9 || (b > 13 && b < 32)) controls++;
        }
        return nul > sample / 100 || controls > sample / 20;
    }

    private String stripRtf(String rtf) {
        if (rtf == null) return "";
        String s = rtf;
        Pattern hex = Pattern.compile("\\\\'([0-9a-fA-F]{2})");
        Matcher m = hex.matcher(s);
        StringBuffer sb = new StringBuffer();
        Charset cp = Charset.forName("windows-1252");
        while (m.find()) {
            int v = Integer.parseInt(m.group(1), 16);
            String ch = new String(new byte[]{(byte) v}, cp);
            m.appendReplacement(sb, Matcher.quoteReplacement(ch));
        }
        m.appendTail(sb);
        s = sb.toString();
        s = s.replaceAll("\\\\par[d]?\\b", "\n");
        s = s.replaceAll("\\\\line\\b", "\n");
        s = s.replaceAll("\\\\tab\\b", "\t");
        s = s.replaceAll("\\\\u(-?\\d+)\\??", " ");
        s = s.replaceAll("\\\\[a-zA-Z]+-?\\d* ?", "");
        s = s.replaceAll("[{}]", "");
        s = s.replace("\\\\", "\\").replace("\\{", "{").replace("\\}", "}");
        return normalizeExtractedText(s);
    }

    private String extractLegacyDoc(byte[] bytes) {
        StringBuilder out = new StringBuilder();

        String utf16 = new String(bytes, StandardCharsets.UTF_16LE);
        Matcher wide = Pattern.compile("[\\p{L}\\p{N}\\p{P}\\p{Zs}\\r\\n\\t]{4,}").matcher(utf16);
        while (wide.find()) appendBlock(out, cleanLegacyChunk(wide.group()));

        String ansi = new String(bytes, Charset.forName("windows-1252"));
        Matcher narrow = Pattern.compile("[\\p{L}\\p{N}\\p{P}\\p{Zs}\\r\\n\\t]{6,}").matcher(ansi);
        while (narrow.find()) {
            String chunk = cleanLegacyChunk(narrow.group());
            if (chunk.length() >= 6 && !chunk.contains("WordDocument") && !chunk.contains("Microsoft Office")) appendBlock(out, chunk);
        }

        String text = normalizeExtractedText(out.toString());
        if (text.replaceAll("\\s+", "").length() < 12) {
            throw new IllegalStateException("DOC legacy: testo non estraibile in modo affidabile da questo file.");
        }
        return text;
    }

    private String cleanLegacyChunk(String s) {
        if (s == null) return "";
        return s.replace('\u0000', ' ').replaceAll("[\\p{Cntrl}&&[^\\r\\n\\t]]", " ").trim();
    }

    private String normalizeExtractedText(String text) {
        if (text == null) return "";
        String s = text.replace("\r\n", "\n").replace('\r', '\n');
        s = s.replace('\u00A0', ' ');
        s = s.replaceAll("[ \\t]+\\n", "\n");
        s = s.replaceAll("\\n[ \\t]+", "\n");
        s = s.replaceAll("\\n{3,}", "\n\n");
        return s.trim();
    }

    private void appendBlock(StringBuilder out, String text) {
        if (text == null) return;
        String t = text.trim();
        if (t.isEmpty()) return;
        if (out.length() > 0) out.append("\n\n");
        out.append(t);
    }

    private static class ZipTextPart implements Comparable<ZipTextPart> {
        final String name;
        final byte[] bytes;
        ZipTextPart(String name, byte[] bytes) {
            this.name = name;
            this.bytes = bytes;
        }
        @Override
        public int compareTo(ZipTextPart other) {
            return naturalName(name).compareTo(naturalName(other.name));
        }
        private static String naturalName(String n) {
            Matcher m = Pattern.compile("(\\d+)").matcher(n);
            StringBuffer out = new StringBuffer();
            while (m.find()) {
                String padded = String.format(Locale.ROOT, "%010d", Long.parseLong(m.group(1)));
                m.appendReplacement(out, padded);
            }
            m.appendTail(out);
            return out.toString();
        }
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
                    String error = "Errore apertura file: " + (e.getMessage() == null ? e.getClass().getSimpleName() : e.getMessage());
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

                try (InputStream raw = getAssets().open("anagram_index.tsv");
                     BufferedReader reader = new BufferedReader(new InputStreamReader(raw, StandardCharsets.UTF_8), 64 * 1024)) {
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

