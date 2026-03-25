package com.legoparser.api

data class Product(
    val id: Long,
    val mcode: String?,
    val barcode: String?,
    val invoice_code: String?,
    val name_ka: String,
    val name_en: String?,
    val image_url: String?,
    val sources: List<String> = emptyList(),
    val created_at: String = "",
    val updated_at: String = ""
)

data class PriceComparison(
    val product_id: Long,
    val mcode: String?,
    val barcode: String?,
    val invoice_code: String?,
    val name_ka: String,
    val name_en: String?,
    val biblusi_xs_price: Double?,
    val biblusi_pepela_price: Double?,
    val wolt_xs_price: Double?,
    val wolt_pepela_price: Double?,
    val glovo_xs_price: Double?,
    val glovo_pepela_price: Double?,
    val wishlist_price: Double?,
    val piccolatoys_price: Double?,
    val kubiki_price: Double?
)

data class ScrapeRun(
    val id: Long,
    val source: String,
    val status: String,
    val trigger_type: String,
    val products_found: Int,
    val products_saved: Int,
    val errors_count: Int,
    val error_log: String?,
    val started_at: String?,
    val completed_at: String?,
    val created_at: String
)

data class ScrapeStatus(
    val is_running: Boolean,
    val runs: List<ScrapeRun>?
)

data class PaginatedResponse<T>(
    val data: List<T>?,
    val total: Int,
    val page: Int,
    val per_page: Int,
    val total_pages: Int
)

data class ScrapeRequest(val source: String)
