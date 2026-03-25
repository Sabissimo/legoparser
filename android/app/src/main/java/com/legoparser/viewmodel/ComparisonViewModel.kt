package com.legoparser.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.legoparser.data.db.AppDatabase
import com.legoparser.data.db.PriceComparisonRow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class ComparisonState(
    val items: List<PriceComparisonRow> = emptyList(),
    val total: Int = 0, val page: Int = 1, val totalPages: Int = 0,
    val search: String = "", val sortBy: String = "name_ka", val sortOrder: String = "asc",
    val isLoading: Boolean = false
)

class ComparisonViewModel(app: Application) : AndroidViewModel(app) {
    private val dao = AppDatabase.get(app).productDao()
    private val _state = MutableStateFlow(ComparisonState())
    val state = _state.asStateFlow()
    private val perPage = 20

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.value = _state.value.copy(isLoading = true)
            val s = _state.value
            val search = s.search.ifBlank { null }
            val total = dao.countComparison(search)
            val offset = (s.page - 1) * perPage
            val items = dao.getComparison(search, s.sortBy, s.sortOrder, perPage, offset)
            _state.value = s.copy(
                items = items, total = total,
                totalPages = (total + perPage - 1) / perPage, isLoading = false
            )
        }
    }

    fun setSearch(q: String) { _state.value = _state.value.copy(search = q, page = 1); load() }
    fun setPage(p: Int) { _state.value = _state.value.copy(page = p); load() }
    fun toggleSort(col: String) {
        val s = _state.value
        val order = if (s.sortBy == col && s.sortOrder == "asc") "desc" else "asc"
        _state.value = s.copy(sortBy = col, sortOrder = order, page = 1); load()
    }
}
