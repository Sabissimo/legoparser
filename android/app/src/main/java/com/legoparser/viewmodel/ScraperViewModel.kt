package com.legoparser.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.legoparser.api.ApiClient
import com.legoparser.api.ScrapeRequest
import com.legoparser.api.ScrapeRun
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class ScraperState(
    val runs: List<ScrapeRun> = emptyList(),
    val isRunning: Boolean = false,
    val isLoading: Boolean = false,
    val error: String? = null
)

class ScraperViewModel : ViewModel() {
    private val _state = MutableStateFlow(ScraperState())
    val state = _state.asStateFlow()

    init { loadRuns(); pollStatus() }

    fun loadRuns() {
        viewModelScope.launch {
            try {
                val runs = ApiClient.api.getScrapeRuns()
                _state.value = _state.value.copy(runs = runs)
            } catch (_: Exception) {}
        }
    }

    private fun pollStatus() {
        viewModelScope.launch {
            while (true) {
                try {
                    val status = ApiClient.api.getScrapeStatus()
                    _state.value = _state.value.copy(isRunning = status.is_running)
                    if (status.is_running) loadRuns()
                } catch (_: Exception) {}
                delay(3000)
            }
        }
    }

    fun startScrape(source: String) {
        viewModelScope.launch {
            try {
                ApiClient.api.startScrape(ScrapeRequest(source))
                _state.value = _state.value.copy(isRunning = true)
                delay(1000)
                loadRuns()
            } catch (e: Exception) {
                _state.value = _state.value.copy(error = e.message)
            }
        }
    }
}
