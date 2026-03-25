package com.legoparser.data.scraper

import android.annotation.SuppressLint
import android.content.Context
import android.webkit.WebView
import android.webkit.WebViewClient
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeout
import org.jsoup.Jsoup
import org.jsoup.nodes.Document

/**
 * Uses Android WebView to load JS-rendered pages and return parsed HTML.
 * Must be called from a context that has access to the main thread for WebView.
 */
object WebViewScraper {
    @SuppressLint("SetJavaScriptEnabled")
    suspend fun fetchRenderedHtml(context: Context, url: String, waitMs: Long = 5000, timeoutMs: Long = 60000): Document {
        return withTimeout(timeoutMs) {
            val deferred = CompletableDeferred<String>()

            withContext(Dispatchers.Main) {
                val webView = WebView(context).apply {
                    settings.javaScriptEnabled = true
                    settings.domStorageEnabled = true
                    settings.userAgentString = "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36"
                    webViewClient = object : WebViewClient() {
                        override fun onPageFinished(view: WebView?, url: String?) {
                            view?.postDelayed({
                                view.evaluateJavascript("document.documentElement.outerHTML") { html ->
                                    val unescaped = html
                                        ?.removeSurrounding("\"")
                                        ?.replace("\\u003C", "<")
                                        ?.replace("\\u003E", ">")
                                        ?.replace("\\\"", "\"")
                                        ?.replace("\\n", "\n")
                                        ?.replace("\\t", "\t")
                                        ?.replace("\\/", "/")
                                        ?: ""
                                    deferred.complete(unescaped)
                                    view.destroy()
                                }
                            }, waitMs)
                        }
                    }
                }
                webView.loadUrl(url)
            }

            Jsoup.parse(deferred.await())
        }
    }
}
