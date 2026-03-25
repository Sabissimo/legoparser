package com.legoparser.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.legoparser.viewmodel.ScraperViewModel
import java.text.SimpleDateFormat
import java.util.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ScraperScreen(vm: ScraperViewModel = viewModel()) {
    val state by vm.state.collectAsStateWithLifecycle()
    var selectedSource by remember { mutableStateOf("all") }
    var expanded by remember { mutableStateOf(false) }
    val sourceLabels = mapOf(
        "all" to "All Sources",
        "biblusi_xs" to "Biblusi XS", "biblusi_pepela" to "Biblusi Pepela",
        "wishlist" to "Wishlist", "piccolatoys" to "PiccolaToys",
        "kubiki" to "Kubiki",
        "wolt_xs" to "Wolt XS", "wolt_pepela" to "Wolt Pepela",
        "glovo_xs" to "Glovo XS", "glovo_pepela" to "Glovo Pepela"
    )

    Column(Modifier.fillMaxSize().padding(16.dp)) {
        Card(Modifier.fillMaxWidth()) {
            Column(Modifier.padding(16.dp)) {
                Text("Start Scrape", style = MaterialTheme.typography.titleMedium)
                Spacer(Modifier.height(12.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }, modifier = Modifier.weight(1f)) {
                        OutlinedTextField(
                            value = sourceLabels[selectedSource] ?: selectedSource,
                            onValueChange = {}, readOnly = true,
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) },
                            modifier = Modifier.menuAnchor()
                        )
                        ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                            sourceLabels.forEach { (value, label) ->
                                DropdownMenuItem(text = { Text(label) }, onClick = { selectedSource = value; expanded = false })
                            }
                        }
                    }
                    Button(onClick = { vm.startScrape(selectedSource) }, enabled = !state.isRunning) {
                        if (state.isRunning) {
                            CircularProgressIndicator(Modifier.size(18.dp), strokeWidth = 2.dp)
                            Spacer(Modifier.width(8.dp)); Text("Running...")
                        } else {
                            Icon(Icons.Default.PlayArrow, null, Modifier.size(18.dp))
                            Spacer(Modifier.width(4.dp)); Text("Start")
                        }
                    }
                }
                if (state.isRunning) { Spacer(Modifier.height(8.dp)); LinearProgressIndicator(Modifier.fillMaxWidth()) }
            }
        }

        Spacer(Modifier.height(16.dp))
        Text("History", style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(8.dp))

        LazyColumn {
            items(state.runs) { run ->
                Card(
                    Modifier.fillMaxWidth().padding(vertical = 2.dp),
                    colors = CardDefaults.cardColors(containerColor = when (run.status) {
                        "completed" -> MaterialTheme.colorScheme.secondaryContainer
                        "running" -> MaterialTheme.colorScheme.primaryContainer
                        "failed" -> MaterialTheme.colorScheme.errorContainer
                        else -> MaterialTheme.colorScheme.surfaceVariant
                    })
                ) {
                    Row(Modifier.padding(10.dp).fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                        Column {
                            Text(run.source.replace("_", " "), style = MaterialTheme.typography.bodyMedium)
                            Text("${run.status} | ${run.productsFound} found, ${run.productsSaved} saved", fontSize = 11.sp)
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            if (run.completedAt != null) {
                                val dur = (run.completedAt!! - run.startedAt) / 1000
                                Text("${dur}s", fontSize = 11.sp)
                            } else if (run.status == "running") Text("...", fontSize = 11.sp)
                            Text(SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date(run.startedAt)), fontSize = 10.sp, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                }
            }
        }
    }
}
