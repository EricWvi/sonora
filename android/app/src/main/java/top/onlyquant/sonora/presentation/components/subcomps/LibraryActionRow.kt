package top.onlyquant.sonora.presentation.components.subcomps

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.SizeTransform
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.animateContentSize
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.animation.expandHorizontally
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.shrinkHorizontally
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.rounded.ArrowBack
import androidx.compose.material.icons.automirrored.rounded.PlaylistAdd
import androidx.compose.material.icons.automirrored.rounded.Sort
import androidx.compose.material.icons.rounded.ChevronRight
import androidx.compose.material.icons.rounded.Home
import androidx.compose.material.icons.rounded.MyLocation
import androidx.compose.material.icons.rounded.Shuffle
import androidx.compose.material.icons.rounded.FilterList
import androidx.compose.material.icons.rounded.Cloud
import androidx.compose.material.icons.rounded.Dataset
import androidx.compose.material.icons.rounded.PhoneAndroid
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExperimentalMaterial3ExpressiveApi
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.FilledTonalIconButton
import androidx.compose.material3.IconButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.MotionScheme
import androidx.compose.material3.PlainTooltip
import androidx.compose.material3.Text
import androidx.compose.material3.TooltipBox
import androidx.compose.material3.TooltipDefaults
import androidx.compose.material3.TooltipAnchorPosition
import androidx.compose.material3.rememberTooltipState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.CompositingStrategy
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.model.StorageFilter
import top.onlyquant.sonora.ui.theme.RoundedSans
import java.io.File
import androidx.compose.ui.res.stringResource

val defaultShape = RoundedCornerShape(26.dp)
val genHeight = 42.dp

// Sonora: without the folders tab's breadcrumbs mode (BREADCRUMBS) and its folder fields.
private enum class LibraryToolbarMode { SHUFFLE, PLAYLIST }

private data class LibraryToolbarContent(
    val isPlaylistTab: Boolean,
    val showImportButton: Boolean,
    val showSortButton: Boolean,
    val showLocateButton: Boolean,
    val showStorageFilterButton: Boolean,
    val currentStorageFilter: StorageFilter,
    val iconRotation: Float
) {
    val mode: LibraryToolbarMode
        get() = when {
            isPlaylistTab -> LibraryToolbarMode.PLAYLIST
            else -> LibraryToolbarMode.SHUFFLE
        }
}

@Composable
fun LibraryActionRow(
    onMainActionClick: () -> Unit,
    iconRotation: Float,
    onSortClick: () -> Unit,
    onLocateClick: () -> Unit = {},
    showSortButton: Boolean,
    showLocateButton: Boolean = false,
    showImportButton: Boolean = true,
    isPlaylistTab: Boolean,
    onImportM3uClick: () -> Unit = {},
    modifier: Modifier = Modifier,
    isShuffleEnabled: Boolean = false,
    showStorageFilterButton: Boolean = false,
    currentStorageFilter: StorageFilter = StorageFilter.ALL,
    onStorageFilterClick: () -> Unit = {}
) {
    // Animate complete arrangements so departing controls keep their tab's labels and width.
    AnimatedContent(
        targetState = LibraryToolbarContent(
            isPlaylistTab = isPlaylistTab,
            showImportButton = showImportButton,
            showSortButton = showSortButton,
            showLocateButton = showLocateButton,
            showStorageFilterButton = showStorageFilterButton,
            currentStorageFilter = currentStorageFilter,
            iconRotation = iconRotation
        ),
        contentKey = { it.mode },
        contentAlignment = Alignment.CenterStart,
        modifier = modifier.fillMaxWidth().height(48.dp),
        transitionSpec = {
            val direction = -1
            ((slideInVertically(tween(240, easing = FastOutSlowInEasing)) { direction * it / 2 } +
                fadeIn(tween(180))) togetherWith
                (slideOutVertically(tween(240, easing = FastOutSlowInEasing)) { -direction * it / 2 } +
                    fadeOut(tween(120)))).using(SizeTransform(clip = false))
        },
        label = "LibraryToolbarMode"
    ) { content ->
        LibraryActionRowContent(
            content = content,
            onMainActionClick = onMainActionClick,
            onSortClick = onSortClick,
            onLocateClick = onLocateClick,
            onImportM3uClick = onImportM3uClick,
            onStorageFilterClick = onStorageFilterClick
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class, ExperimentalMaterial3ExpressiveApi::class)
@Composable
private fun LibraryActionRowContent(
    content: LibraryToolbarContent,
    onMainActionClick: () -> Unit,
    onSortClick: () -> Unit,
    onLocateClick: () -> Unit,
    onImportM3uClick: () -> Unit,
    onStorageFilterClick: () -> Unit
) {
    val isPlaylistTab = content.isPlaylistTab
    val showSortButton = content.showSortButton
    val showLocateButton = content.showLocateButton
    val showStorageFilterButton = content.showStorageFilterButton
    val currentStorageFilter = content.currentStorageFilter
    val iconRotation = content.iconRotation
    val shouldShowImport = isPlaylistTab && content.showImportButton

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(48.dp)
            .padding(start = 4.dp),
        horizontalArrangement = Arrangement.Absolute.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(modifier = Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.CenterStart) {
            run {
                val outerCorner = 26.dp
                val innerCorner = 8.dp

                val newButtonEndCorner by animateDpAsState(
                    targetValue = if (shouldShowImport) innerCorner else outerCorner,
                    label = "NewButtonEndCorner"
                )

                val importButtonStartCorner by animateDpAsState(
                    targetValue = innerCorner,
                    label = "ImportButtonStartCorner"
                )
                Row(verticalAlignment = Alignment.CenterVertically) {
                    val buttonContainerColor = MaterialTheme.colorScheme.tertiaryContainer
                    val buttonContentColor = MaterialTheme.colorScheme.onTertiaryContainer
                    
                    FilledTonalButton(
                        onClick = onMainActionClick,
                        shape = RoundedCornerShape(
                            topStart = 26.dp, bottomStart = 26.dp,
                            topEnd =  newButtonEndCorner, bottomEnd = newButtonEndCorner
                        ),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = buttonContainerColor,
                            contentColor = buttonContentColor
                        ),
                        elevation = ButtonDefaults.buttonElevation(defaultElevation = 4.dp, pressedElevation = 6.dp),
                        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 10.dp),
                        modifier = Modifier.height(genHeight)
                    ) {
                        val icon = if (isPlaylistTab) Icons.AutoMirrored.Rounded.PlaylistAdd else Icons.Rounded.Shuffle
                        val text = if (isPlaylistTab) {
                            stringResource(R.string.library_action_new)
                        } else {
                            stringResource(R.string.shortcut_shuffle_short)
                        }
                        val contentDesc = if (isPlaylistTab) {
                            stringResource(R.string.cd_create_new_playlist)
                        } else {
                            stringResource(R.string.cd_shuffle_play)
                        }

                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            Icon(
                                imageVector = icon,
                                contentDescription = contentDesc,
                                modifier = Modifier.size(20.dp).rotate(iconRotation)
                            )
                            Text(
                                modifier = Modifier.animateContentSize(),
                                text = text,
                                overflow = TextOverflow.Ellipsis,
                                style = MaterialTheme.typography.labelLarge,
                                fontWeight = FontWeight.Medium
                            )
                        }
                    }

                    AnimatedVisibility(
                        visible = shouldShowImport,
                        enter = fadeIn() + expandHorizontally(
                            expandFrom = Alignment.Start,
                            clip = false,
                            animationSpec = spring(
                                dampingRatio = Spring.DampingRatioMediumBouncy,
                                stiffness = Spring.StiffnessLow
                            )
                        ),
                        exit = fadeOut() + shrinkHorizontally(
                            shrinkTowards = Alignment.Start,
                            clip = false,
                            animationSpec = spring(
                                dampingRatio = Spring.DampingRatioNoBouncy,
                                stiffness = Spring.StiffnessMedium
                            )
                        )
                    ) {
                        Row(modifier = Modifier.height(genHeight), verticalAlignment = Alignment.CenterVertically) {
                            Spacer(modifier = Modifier.width(8.dp))

                            FilledTonalButton(
                                onClick = onImportM3uClick,
                                shape = RoundedCornerShape(
                                    topStart = importButtonStartCorner,
                                    bottomStart = importButtonStartCorner,
                                    topEnd = 26.dp,
                                    bottomEnd = 26.dp
                                ),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = MaterialTheme.colorScheme.secondaryContainer,
                                    contentColor = MaterialTheme.colorScheme.onSecondaryContainer
                                ),
                                elevation = ButtonDefaults.buttonElevation(
                                    defaultElevation = 4.dp,
                                    pressedElevation = 6.dp
                                ),
                                contentPadding = PaddingValues(
                                    horizontal = 14.dp,
                                    vertical = 10.dp
                                ),
                                modifier = Modifier.height(genHeight)
                            ) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    Icon(
                                        painter = painterResource(R.drawable.rounded_upload_file_24),
                                        contentDescription = stringResource(R.string.cd_import_m3u_playlist),
                                        modifier = Modifier.size(20.dp)
                                    )
                                    Text(
                                        text = stringResource(R.string.import_file),
                                        overflow = TextOverflow.Ellipsis,
                                        style = MaterialTheme.typography.labelLarge,
                                        fontWeight = FontWeight.Medium
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }


        Spacer(modifier = Modifier.width(8.dp))

        if (showSortButton) {
            val outerCorner = 26.dp
            
            val sortStartCorner by animateDpAsState(
                targetValue = if (showLocateButton || showStorageFilterButton) 8.dp else outerCorner,
                label = "SortStartCorner"
            )

            val filterEndCorner = 8.dp
            val filterStartCorner by animateDpAsState(
                targetValue = if (showLocateButton) 8.dp else outerCorner,
                label = "FilterStartCorner"
            )
            
            val locateEndCorner = 8.dp

            val gapBetweenLocateAndNext by animateDpAsState(
                targetValue = if (showLocateButton) 4.dp else 0.dp,
                label = "GapLocate"
            )
            val gapBetweenFilterAndSort by animateDpAsState(
                targetValue = if (showStorageFilterButton) 4.dp else 0.dp,
                label = "GapFilter"
            )


            Row(verticalAlignment = Alignment.CenterVertically) {
                AnimatedVisibility(
                    visible = showLocateButton,
                    enter = slideInHorizontally(initialOffsetX = { it / 2 }) + fadeIn(),
                    exit = slideOutHorizontally(targetOffsetX = { it / 2 }) + fadeOut()
                ) {
                    FilledTonalIconButton(
                        onClick = onLocateClick,
                        shape = RoundedCornerShape(
                            topStart = outerCorner,
                            bottomStart = outerCorner,
                            topEnd = locateEndCorner,
                            bottomEnd = locateEndCorner
                        ),
                        modifier = Modifier.size(genHeight)
                    ) {
                        Icon(
                            imageVector = Icons.Rounded.MyLocation,
                            contentDescription = stringResource(R.string.cd_locate_current_song)
                        )
                    }
                }
                
                Spacer(modifier = Modifier.width(gapBetweenLocateAndNext))

                AnimatedVisibility(
                    visible = showStorageFilterButton,
                    enter = slideInHorizontally(initialOffsetX = { it / 2 }) + fadeIn(),
                    exit = slideOutHorizontally(targetOffsetX = { it / 2 }) + fadeOut()
                ) {
                     val finalIcon = when(currentStorageFilter) {
                         StorageFilter.ALL -> Icons.Rounded.Dataset
                         StorageFilter.ONLINE -> Icons.Rounded.Cloud
                         StorageFilter.OFFLINE -> Icons.Rounded.PhoneAndroid
                     }
                     val tooltipText = when(currentStorageFilter) {
                         StorageFilter.ALL -> stringResource(R.string.library_storage_filter_all_songs)
                         StorageFilter.ONLINE -> stringResource(R.string.library_storage_filter_online)
                         StorageFilter.OFFLINE -> stringResource(R.string.library_storage_filter_offline)
                     }
                     val tooltipState = rememberTooltipState()

                    TooltipBox(
                        positionProvider = TooltipDefaults.rememberTooltipPositionProvider(TooltipAnchorPosition.Above),
                        tooltip = {
                            PlainTooltip {
                                Text(tooltipText)
                            }
                        },
                        state = tooltipState
                    ) {
                        // The button squishes and rounds out while held, the icon spins in when
                        // the source changes, and a source other than All keeps a tertiary tint.
                        val filterMotion = remember { MotionScheme.expressive() }
                        val filterInteraction = remember { MutableInteractionSource() }
                        val filterPressed by filterInteraction.collectIsPressedAsState()
                        val filterScale by animateFloatAsState(
                            targetValue = if (filterPressed) 0.88f else 1f,
                            animationSpec = filterMotion.fastSpatialSpec(),
                            label = "StorageFilterScale"
                        )
                        val pressedCorner = genHeight / 2
                        val filterStart by animateDpAsState(
                            targetValue = if (filterPressed) pressedCorner else filterStartCorner,
                            animationSpec = filterMotion.fastSpatialSpec(),
                            label = "StorageFilterStartCorner"
                        )
                        val filterEnd by animateDpAsState(
                            targetValue = if (filterPressed) pressedCorner else filterEndCorner,
                            animationSpec = filterMotion.fastSpatialSpec(),
                            label = "StorageFilterEndCorner"
                        )
                        val isFiltering = currentStorageFilter != StorageFilter.ALL
                        val filterContainer by animateColorAsState(
                            targetValue = if (isFiltering) MaterialTheme.colorScheme.tertiaryContainer
                            else MaterialTheme.colorScheme.secondaryContainer,
                            animationSpec = filterMotion.defaultEffectsSpec(),
                            label = "StorageFilterContainer"
                        )
                        val filterContent by animateColorAsState(
                            targetValue = if (isFiltering) MaterialTheme.colorScheme.onTertiaryContainer
                            else MaterialTheme.colorScheme.onSecondaryContainer,
                            animationSpec = filterMotion.defaultEffectsSpec(),
                            label = "StorageFilterContent"
                        )
                        FilledTonalIconButton(
                            onClick = onStorageFilterClick,
                            shape = RoundedCornerShape(
                                topStart = filterStart,
                                bottomStart = filterStart,
                                topEnd = filterEnd,
                                bottomEnd = filterEnd
                            ),
                            colors = IconButtonDefaults.filledTonalIconButtonColors(
                                containerColor = filterContainer,
                                contentColor = filterContent
                            ),
                            interactionSource = filterInteraction,
                            modifier = Modifier
                                .size(genHeight)
                                .graphicsLayer {
                                    scaleX = filterScale
                                    scaleY = filterScale
                                }
                        ) {
                            AnimatedContent(
                                targetState = finalIcon,
                                transitionSpec = {
                                    (scaleIn(filterMotion.fastSpatialSpec(), initialScale = 0.4f) +
                                        fadeIn(filterMotion.fastEffectsSpec())) togetherWith
                                        (scaleOut(filterMotion.fastSpatialSpec(), targetScale = 0.4f) +
                                            fadeOut(filterMotion.fastEffectsSpec()))
                                },
                                label = "StorageFilterIcon"
                            ) { icon ->
                                val spin = remember { Animatable(-90f) }
                                LaunchedEffect(Unit) { spin.animateTo(0f, filterMotion.defaultSpatialSpec()) }
                                Icon(
                                    imageVector = icon,
                                    contentDescription = tooltipText,
                                    modifier = Modifier.graphicsLayer { rotationZ = spin.value }
                                )
                            }
                        }
                    }
                }

                Spacer(modifier = Modifier.width(gapBetweenFilterAndSort))

                FilledTonalIconButton(
                    onClick = onSortClick,
                    shape = RoundedCornerShape(
                        topStart = sortStartCorner,
                        bottomStart = sortStartCorner,
                        topEnd = outerCorner,
                        bottomEnd = outerCorner
                    ),
                    modifier = Modifier.size(genHeight)
                ) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Rounded.Sort,
                        contentDescription = stringResource(R.string.cd_sort_options),
                    )
                }
            }
        }
    }
}

// Sonora: the original's Breadcrumbs (the folders tab's path) is left out; Sonora has no folders tab.
