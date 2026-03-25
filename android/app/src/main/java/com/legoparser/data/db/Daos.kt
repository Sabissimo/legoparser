package com.legoparser.data.db

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Dao
interface ProductDao {
    @Query("""
        SELECT p.id, p.mcode, p.barcode, p.invoice_code, p.name_ka, p.name_en, p.image_url,
               GROUP_CONCAT(DISTINCT pr.source) as sources
        FROM products p LEFT JOIN prices pr ON pr.product_id = p.id
        WHERE (:search IS NULL OR p.name_ka LIKE '%' || :search || '%' OR p.invoice_code LIKE '%' || :search || '%' OR p.mcode LIKE '%' || :search || '%' OR p.barcode LIKE '%' || :search || '%')
        GROUP BY p.id
        ORDER BY
            CASE WHEN :sortBy = 'name_ka' AND :sortOrder = 'asc' THEN p.name_ka END ASC,
            CASE WHEN :sortBy = 'name_ka' AND :sortOrder = 'desc' THEN p.name_ka END DESC,
            CASE WHEN :sortBy = 'invoice_code' AND :sortOrder = 'asc' THEN p.invoice_code END ASC,
            CASE WHEN :sortBy = 'invoice_code' AND :sortOrder = 'desc' THEN p.invoice_code END DESC,
            CASE WHEN :sortBy = 'mcode' AND :sortOrder = 'asc' THEN p.mcode END ASC,
            CASE WHEN :sortBy = 'mcode' AND :sortOrder = 'desc' THEN p.mcode END DESC
        LIMIT :limit OFFSET :offset
    """)
    suspend fun getProducts(search: String?, sortBy: String, sortOrder: String, limit: Int, offset: Int): List<ProductWithSources>

    @Query("SELECT COUNT(DISTINCT p.id) FROM products p WHERE (:search IS NULL OR p.name_ka LIKE '%' || :search || '%' OR p.invoice_code LIKE '%' || :search || '%' OR p.mcode LIKE '%' || :search || '%')")
    suspend fun countProducts(search: String?): Int

    @Query("SELECT id FROM products WHERE invoice_code = :code LIMIT 1")
    suspend fun findByInvoiceCode(code: String): Long?

    @Query("SELECT id FROM products WHERE barcode = :barcode LIMIT 1")
    suspend fun findByBarcode(barcode: String): Long?

    @Query("SELECT id FROM products WHERE name_ka = :name LIMIT 1")
    suspend fun findByName(name: String): Long?

    @Insert
    suspend fun insert(product: ProductEntity): Long

    @Query("UPDATE products SET mcode = COALESCE(:mcode, mcode), barcode = COALESCE(:barcode, barcode), invoice_code = COALESCE(:invoiceCode, invoice_code), image_url = COALESCE(:imageUrl, image_url), updated_at = :now WHERE id = :id")
    suspend fun update(id: Long, mcode: String?, barcode: String?, invoiceCode: String?, imageUrl: String?, now: Long = System.currentTimeMillis())

    @Query("""
        SELECT p.id as product_id, p.mcode, p.barcode, p.invoice_code, p.name_ka, p.name_en,
            MAX(CASE WHEN lp.source = 'biblusi_xs' THEN COALESCE(lp.discounted_price, lp.original_price) END) as biblusi_xs,
            MAX(CASE WHEN lp.source = 'biblusi_pepela' THEN COALESCE(lp.discounted_price, lp.original_price) END) as biblusi_pepela,
            MAX(CASE WHEN lp.source = 'wolt_xs' THEN COALESCE(lp.discounted_price, lp.original_price) END) as wolt_xs,
            MAX(CASE WHEN lp.source = 'wolt_pepela' THEN COALESCE(lp.discounted_price, lp.original_price) END) as wolt_pepela,
            MAX(CASE WHEN lp.source = 'glovo_xs' THEN COALESCE(lp.discounted_price, lp.original_price) END) as glovo_xs,
            MAX(CASE WHEN lp.source = 'glovo_pepela' THEN COALESCE(lp.discounted_price, lp.original_price) END) as glovo_pepela,
            MAX(CASE WHEN lp.source = 'wishlist' THEN COALESCE(lp.discounted_price, lp.original_price) END) as wishlist,
            MAX(CASE WHEN lp.source = 'piccolatoys' THEN COALESCE(lp.discounted_price, lp.original_price) END) as piccolatoys,
            MAX(CASE WHEN lp.source = 'kubiki' THEN COALESCE(lp.discounted_price, lp.original_price) END) as kubiki
        FROM products p LEFT JOIN prices lp ON lp.product_id = p.id
        WHERE (:search IS NULL OR p.name_ka LIKE '%' || :search || '%' OR p.invoice_code LIKE '%' || :search || '%')
        GROUP BY p.id
        ORDER BY
            CASE WHEN :sortBy = 'name_ka' AND :sortOrder = 'asc' THEN p.name_ka END ASC,
            CASE WHEN :sortBy = 'name_ka' AND :sortOrder = 'desc' THEN p.name_ka END DESC,
            CASE WHEN :sortBy = 'invoice_code' AND :sortOrder = 'asc' THEN p.invoice_code END ASC,
            CASE WHEN :sortBy = 'invoice_code' AND :sortOrder = 'desc' THEN p.invoice_code END DESC
        LIMIT :limit OFFSET :offset
    """)
    suspend fun getComparison(search: String?, sortBy: String, sortOrder: String, limit: Int, offset: Int): List<PriceComparisonRow>

    @Query("SELECT COUNT(DISTINCT p.id) FROM products p WHERE (:search IS NULL OR p.name_ka LIKE '%' || :search || '%' OR p.invoice_code LIKE '%' || :search || '%')")
    suspend fun countComparison(search: String?): Int
}

@Dao
interface PriceDao {
    @Insert
    suspend fun insert(price: PriceEntity): Long

    @Query("DELETE FROM prices WHERE source = :source")
    suspend fun deleteBySource(source: String)
}

@Dao
interface ScrapeRunDao {
    @Insert
    suspend fun insert(run: ScrapeRunEntity): Long

    @Update
    suspend fun update(run: ScrapeRunEntity)

    @Query("SELECT * FROM scrape_runs ORDER BY started_at DESC LIMIT 30")
    fun getAll(): Flow<List<ScrapeRunEntity>>

    @Query("SELECT COUNT(*) > 0 FROM scrape_runs WHERE status = 'running'")
    fun isRunning(): Flow<Boolean>
}
