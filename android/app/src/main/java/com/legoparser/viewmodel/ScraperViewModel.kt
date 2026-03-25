package com.legoparser.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.legoparser.data.db.AppDatabase
import com.legoparser.data.db.ScrapeRunEntity
import com.legoparser.data.scraper.ScraperManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

data class ScraperState(
    val runs: List<ScrapeRunEntity> = emptyList(),
    val isRunning: Boolean = false
)

class ScraperViewModel(app: Application) : AndroidViewModel(app) {
    private val db = AppDatabase.get(app)
    private val manager = ScraperManager(app)

    val state: StateFlow<ScraperState> = combine(
        db.scrapeRunDao().getAll(),
        db.scrapeRunDao().isRunning()
    ) { runs, running ->
        ScraperState(runs = runs, isRunning = running)
    }.stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), ScraperState())

    val sources = manager.allSources

    fun startScrape(source: String) {
        viewModelScope.launch(Dispatchers.IO) {
            if (source == "all") manager.runAll()
            else manager.runSource(source)
        }
    }
}
