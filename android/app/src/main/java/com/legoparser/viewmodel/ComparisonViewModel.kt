package com.legoparser.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.legoparser.api.ApiClient
import com.legoparser.api.PriceComparison
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class ComparisonState(
    val items: List<PriceComparison> = emptyList(),
    val total: Int = 0,
    val page: Int = 1,
    val totalPages: Int = 0,
    val search: String = "",
    val sortBy: String = "name_ka",
    val sortOrder: String = "asc",
    val isLoading: Boolean = false,
    val error: String? = null
)

class ComparisonViewModel : ViewModel() {
    private val _state = MutableStateFlow(ComparisonState())
    val state = _state.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.value = _state.value.copy(isLoading = true, error = null)
            try {
                val s = _state.value
                val resp = ApiClient.api.getComparison(
                    page = s.page, search = s.search.ifBlank { null },
                    sortBy = s.sortBy, sortOrder = s.sortOrder
                )
                _state.value = s.copy(
                    items = resp.data ?: emptyList(),
                    total = resp.total, totalPages = resp.total_pages,
                    isLoading = false
                )
            } catch (e: Exception) {
                _state.value = _state.value.copy(isLoading = false, error = e.message)
            }
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
