package top.onlyquant.sonora.presentation.components

import android.app.Activity
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.animateContentSize
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.wrapContentHeight
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.QueueMusic
import androidx.compose.material.icons.automirrored.rounded.PlaylistAdd
import androidx.compose.material.icons.automirrored.rounded.QueueMusic
import androidx.compose.material.icons.rounded.Delete
import androidx.compose.material.icons.rounded.CloudDownload
import androidx.compose.material.icons.rounded.Edit
import androidx.compose.material.icons.rounded.Favorite
import androidx.compose.material.icons.rounded.FavoriteBorder
import androidx.compose.material.icons.rounded.FolderZip
import androidx.compose.material.icons.rounded.HeartBroken
import androidx.compose.material.icons.rounded.PlayArrow
import androidx.compose.material.icons.rounded.Share
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExperimentalMaterial3ExpressiveApi
import androidx.compose.material3.FilledIconButton
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.FilledTonalIconButton
import androidx.compose.material3.FloatingActionButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.zIndex
import coil.size.Size
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.ui.theme.RoundedSans
import racra.compose.smooth_corner_rect_library.AbsoluteSmoothCornerShape

/**
 * Bottom sheet for batch operations on multiple selected songs.
 * Modeled after SongInfoBottomSheet but without the Info/Edit tabs.
 *
 * @param selectedSongs List of selected songs in selection order
 * @param favoriteSongIds Set of song IDs that are currently favorited
 * @param onDismiss Callback when sheet is dismissed
 * @param onPlayAll Play all selected songs
 * @param onAddToQueue Add all to end of queue
 * @param onPlayNext Add all to play next
 * @param onAddToPlaylist Open playlist picker for batch add
 * @param onToggleLikeAll Toggle like status - if all are liked, unlike all; otherwise like all
 */
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import top.onlyquant.sonora.R
import top.onlyquant.sonora.presentation.components.subcomps.AutoSizingTextToFill
import top.onlyquant.sonora.presentation.components.subcomps.TightWrapText
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.toImmutableList

@OptIn(ExperimentalMaterial3Api::class, ExperimentalMaterial3ExpressiveApi::class)
/**
 * From MultiSelectionBottomSheet in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Sonora: play all,
 * like, add to queue, play next and add to playlist (Eric, 2026-11-11); without share, download,
 * delete and batch edit.
 */
@Composable
fun MultiSelectionBottomSheet(
    selectedSongs: ImmutableList<Song>,
    favoriteSongIds: Set<String> = emptySet(),
    onDismiss: () -> Unit,
    onPlayAll: () -> Unit,
    onAddToQueue: () -> Unit,
    onPlayNext: () -> Unit,
    onAddToPlaylist: () -> Unit,
    onToggleLikeAll: (shouldLike: Boolean) -> Unit
) {
    val context = LocalContext.current
    val sheetState = rememberModalSheetState(skipPartiallyExpanded = true)
    
    val allAreLiked by remember(selectedSongs, favoriteSongIds) {
        derivedStateOf {
            selectedSongs.isNotEmpty() && selectedSongs.all { favoriteSongIds.contains(it.id) }
        }
    }
    
    val evenCornerRadius = 26.dp
    val buttonShape = AbsoluteSmoothCornerShape(
        cornerRadiusTR = evenCornerRadius, smoothnessAsPercentBR = 60, 
        cornerRadiusBR = evenCornerRadius, smoothnessAsPercentTL = 60, 
        cornerRadiusTL = evenCornerRadius, smoothnessAsPercentBL = 60,
        cornerRadiusBL = evenCornerRadius, smoothnessAsPercentTR = 60
    )

    val favoriteButtonCornerRadius by animateDpAsState(
        targetValue = if (allAreLiked) evenCornerRadius else 60.dp,
        animationSpec = tween(durationMillis = 300), label = "FavoriteCornerAnimation"
    )
    val favoriteButtonContainerColor by animateColorAsState(
        targetValue = if (allAreLiked) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant,
        animationSpec = tween(durationMillis = 300), label = "FavoriteContainerColorAnimation"
    )
    val favoriteButtonContentColor by animateColorAsState(
        targetValue = if (allAreLiked) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant,
        animationSpec = tween(durationMillis = 300), label = "FavoriteContentColorAnimation"
    )

    val favoriteButtonShape = remember(favoriteButtonCornerRadius) {
        AbsoluteSmoothCornerShape(
            cornerRadiusTR = favoriteButtonCornerRadius, smoothnessAsPercentBR = 60, cornerRadiusBR = favoriteButtonCornerRadius,
            smoothnessAsPercentTL = 60, cornerRadiusTL = favoriteButtonCornerRadius, smoothnessAsPercentBL = 60,
            cornerRadiusBL = favoriteButtonCornerRadius, smoothnessAsPercentTR = 60
        )
    }
    
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState
    ) {
        Box(
            modifier = Modifier
                .animateContentSize(animationSpec = tween(durationMillis = 200))
                .fillMaxWidth(),
            contentAlignment = Alignment.TopCenter
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp)
            ) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(0.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    val stackedImageSize = 66.dp
                    val stackedOverlap = 33.dp
                    val stackedCount = selectedSongs.take(4).size
                    val stackedWidth = if (stackedCount > 0) {
                        (stackedImageSize - stackedOverlap) * (stackedCount - 1) + stackedImageSize
                    } else 0.dp
                    
                    StackedAlbumArts(
                        songs = remember(selectedSongs) { selectedSongs.take(4).toImmutableList() },
                        modifier = Modifier
                            .height(74.dp)
                            .width(stackedWidth)
                    )
                    
                    Spacer(modifier = Modifier.width(10.dp))
                    
                    Column(
                        modifier = Modifier
                            .weight(1f)
                            .wrapContentHeight(),
                        verticalArrangement = Arrangement.Center
                    ){
                        AutoSizingTextToFill(
                            text = stringResource(R.string.multi_selection_songs_count_upper, selectedSongs.size),
                            style = MaterialTheme.typography.headlineSmall,
                            fontWeight = FontWeight.Bold,
                            fontFamily = RoundedSans,
                            modifier = Modifier.padding(end = 4.dp),
                            minFontSize = 16.sp,
                            maxFontSizeLimit = 24.sp,
                            maxLines = 2,
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(
                            text = stringResource(R.string.multi_selection_selected),
                            style = MaterialTheme.typography.bodyLarge,
                            fontFamily = RoundedSans,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            overflow = TextOverflow.Ellipsis,
                            maxLines = 1,
                        )
                    }

                    // Sonora: the original has a batch edit button here (Eric, 2026-11-11: left out).
                }
                
                Spacer(modifier = Modifier.height(16.dp))
                
                LazyColumn(
                    modifier = Modifier.fillMaxWidth(),
                    contentPadding = PaddingValues(bottom = 32.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    item {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(IntrinsicSize.Min),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(10.dp)
                        ) {
                            FilledTonalButton(
                                modifier = Modifier
                                    .weight(0.5f)
                                    .heightIn(min = 80.dp),
                                colors = ButtonDefaults.filledTonalButtonColors(
                                    containerColor = MaterialTheme.colorScheme.primaryContainer,
                                    contentColor = MaterialTheme.colorScheme.onPrimaryContainer
                                ),
                                contentPadding = PaddingValues(horizontal = 10.dp),
                                shape = buttonShape,
                                onClick = {
                                    onPlayAll()
                                    onDismiss()
                                }
                            ) {
                                Icon(
                                    Icons.Rounded.PlayArrow,
                                    contentDescription = stringResource(R.string.cd_play_all)
                                )
                                Spacer(Modifier.width(6.dp))
                                TightWrapText(
                                    text = stringResource(R.string.action_play_all),
                                    modifier = Modifier.padding(end = 4.dp),
                                    overflow = TextOverflow.Ellipsis,
                                    maxLines = 2,
                                    lineHeight = 22.sp,
                                    style = MaterialTheme.typography.titleMediumEmphasized
                                )
                            }
                            // Sonora: without the share button, play all and like split the row evenly and
                            // like is a labelled pill like play all (Eric, 2026-11-11); the original has
                            // play all 0.5, an icon-only like 0.25 and share 0.25.
                            FilledTonalButton(
                                modifier = Modifier
                                    .weight(0.5f)
                                    .heightIn(min = 80.dp),
                                colors = ButtonDefaults.filledTonalButtonColors(
                                    containerColor = favoriteButtonContainerColor,
                                    contentColor = favoriteButtonContentColor
                                ),
                                contentPadding = PaddingValues(horizontal = 10.dp),
                                shape = favoriteButtonShape,
                                onClick = {
                                    onToggleLikeAll(!allAreLiked)
                                    onDismiss()
                                }
                            ) {
                                Icon(
                                    imageVector = if (allAreLiked)
                                        Icons.Rounded.HeartBroken
                                    else
                                        Icons.Rounded.FavoriteBorder,
                                    contentDescription = stringResource(
                                        if (allAreLiked) R.string.cd_unlike_all else R.string.cd_like_all
                                    )
                                )
                                Spacer(Modifier.width(6.dp))
                                TightWrapText(
                                    text = stringResource(R.string.action_like),
                                    modifier = Modifier.padding(end = 4.dp),
                                    overflow = TextOverflow.Ellipsis,
                                    maxLines = 2,
                                    lineHeight = 22.sp,
                                    style = MaterialTheme.typography.titleMediumEmphasized
                                )
                            }
                            // Sonora: the original has "share all as zip" here (Eric, 2026-11-11: left out).
                        }
                    }

                    // Sonora: the original offers downloading the selected cloud songs here; caching comes with step 7.
                    
                    item {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(IntrinsicSize.Min),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(10.dp)
                        ) {
                            FilledTonalButton(
                                modifier = Modifier
                                    .weight(0.6f)
                                    .heightIn(min = 66.dp),
                                colors = ButtonDefaults.filledTonalButtonColors(
                                    containerColor = MaterialTheme.colorScheme.tertiaryContainer,
                                    contentColor = MaterialTheme.colorScheme.onTertiaryContainer
                                ),
                                contentPadding = PaddingValues(horizontal = 10.dp),
                                shape = CircleShape,
                                onClick = {
                                    onAddToQueue()
                                    onDismiss()
                                }
                            ) {
                                Icon(
                                    Icons.AutoMirrored.Rounded.QueueMusic,
                                    contentDescription = stringResource(R.string.cd_add_all_to_queue)
                                )
                                Spacer(Modifier.width(6.dp))
                                TightWrapText(
                                    text = stringResource(R.string.action_add_to_queue),
                                    modifier = Modifier.padding(end = 4.dp),
                                    overflow = TextOverflow.Ellipsis,
                                    maxLines = 2,
                                    lineHeight = 20.sp
                                )
                            }
                            
                            FilledTonalButton(
                                modifier = Modifier
                                    .weight(0.4f)
                                    .heightIn(min = 66.dp),
                                colors = ButtonDefaults.filledTonalButtonColors(
                                    containerColor = MaterialTheme.colorScheme.tertiary,
                                    contentColor = MaterialTheme.colorScheme.onTertiary
                                ),
                                contentPadding = PaddingValues(horizontal = 10.dp),
                                shape = CircleShape,
                                onClick = {
                                    onPlayNext()
                                    onDismiss()
                                }
                            ) {
                                Icon(
                                    Icons.AutoMirrored.Filled.QueueMusic,
                                    contentDescription = stringResource(R.string.cd_play_next_in_queue)
                                )
                                Spacer(Modifier.width(6.dp))
                                TightWrapText(
                                    text = stringResource(R.string.action_queue_next),
                                    modifier = Modifier.padding(end = 4.dp),
                                    overflow = TextOverflow.Ellipsis,
                                    maxLines = 2,
                                    lineHeight = 20.sp
                                )
                            }
                        }
                    }
                    
                    item {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(IntrinsicSize.Min),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(10.dp)
                        ) {
                            FilledTonalButton(
                                modifier = Modifier
                                    .weight(0.5f)
                                    .heightIn(min = 66.dp),
                                colors = ButtonDefaults.filledTonalButtonColors(
                                    containerColor = MaterialTheme.colorScheme.secondaryContainer,
                                    contentColor = MaterialTheme.colorScheme.onSecondaryContainer
                                ),
                                contentPadding = PaddingValues(horizontal = 10.dp),
                                shape = CircleShape,
                                onClick = {
                                    onAddToPlaylist()
                                    onDismiss()
                                }
                            ) {
                                Icon(
                                    Icons.AutoMirrored.Rounded.PlaylistAdd,
                                    contentDescription = stringResource(R.string.cd_add_to_playlist)
                                )
                                Spacer(Modifier.width(6.dp))
                                TightWrapText(
                                    text = stringResource(R.string.shortcut_playlist_short),
                                    modifier = Modifier.padding(end = 4.dp),
                                    overflow = TextOverflow.Ellipsis,
                                    maxLines = 2,
                                    lineHeight = 20.sp
                                )
                            }
                            // Sonora: the original has "delete all" next to it (Eric, 2026-11-11: left out).
                        }
                    }
                }
            }
        }
    }
}

/**
 * Displays stacked album art images with overlap effect for the bottom sheet header.
 * Uses CircleShape with a border matching the sheet background color.
 */
@Composable
private fun StackedAlbumArts(
    songs: ImmutableList<Song>,
    modifier: Modifier = Modifier
) {
    val imageSize = 66.dp
    val overlap = 33.dp
    val borderWidth = 3.dp
    val borderColor = MaterialTheme.colorScheme.surfaceContainerLow
    
    Box(
        modifier = modifier,
        contentAlignment = Alignment.CenterStart
    ) {
        songs.forEachIndexed { index, song ->
            val offsetX = index * (imageSize.value - overlap.value)
            
            Box(
                modifier = Modifier
                    .offset { IntOffset(offsetX.dp.roundToPx(), 0) }
                    .zIndex((songs.size - index).toFloat())
                    .size(imageSize)
                    .background(borderColor, CircleShape)
            ) {
                Box(
                    modifier = Modifier
                        .matchParentSize()
                        .padding(borderWidth)
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.surfaceVariant)
                ) {
                    SmartImage(
                        model = song.albumArtUriString,
                        contentDescription = song.title,
                        shape = CircleShape,
                        targetSize = Size(168, 168),
                        modifier = Modifier.matchParentSize()
                    )
                }
            }
        }
    }
}
