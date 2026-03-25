package com.legoparser.data.scraper

data class ScrapedProduct(
    val mcode: String? = null,
    val barcode: String? = null,
    val invoiceCode: String? = null,
    val nameKa: String,
    val nameEn: String? = null,
    val imageUrl: String? = null,
    val originalPrice: Double? = null,
    val discountPercent: Double? = null,
    val discountedPrice: Double? = null,
    val inStock: Boolean = true,
    val sourceUrl: String? = null
)

// Extract LEGO set number (4-6 digits) from product name
private val startRegex = Regex("""^\s*(\d{4,6})\b""")
private val afterLegoRegex = Regex("""(?i)(?:lego|ლეგო)\s+(\d{4,6})\b""")
private val endRegex = Regex("""\b(\d{4,6})\s*$""")
private val anywhereRegex = Regex("""\b(\d{5,6})\b""")

fun extractLegoID(name: String): String? {
    val clean = name.trim().replace("-", " ")
    startRegex.find(clean)?.groupValues?.get(1)?.let { return it }
    afterLegoRegex.find(clean)?.groupValues?.get(1)?.let { return it }
    endRegex.find(clean)?.groupValues?.get(1)?.let { return it }
    return null
}

fun extractLegoIDAnywhere(name: String): String? {
    extractLegoID(name)?.let { return it }
    val clean = name.trim().replace("-", " ")
    anywhereRegex.find(clean)?.groupValues?.get(1)?.let { return it }
    return null
}

fun isLegoProduct(name: String): Boolean {
    val lower = name.lowercase()
    return "lego" in lower || "ლეგო" in lower
}

fun parsePrice(text: String): Double? {
    val clean = text.trim().replace(",", ".").replace(Regex("[^0-9.]"), "")
    return clean.toDoubleOrNull()
}

fun calcDiscountPercent(original: Double, discounted: Double): Double? {
    if (original <= 0 || discounted >= original) return null
    return (1 - discounted / original) * 100
}
