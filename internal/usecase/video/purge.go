package video

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/pkg/apperror"
)

// PurgeOrphansInput controls which orphan directions are checked and whether
// to commit deletions.
type PurgeOrphansInput struct {
	// IncludeS3Orphans also scans S3 for objects not referenced by any DB row.
	IncludeS3Orphans bool
	// DryRun reports orphans without performing any deletions.
	DryRun bool
}

// PurgeOrphansOutput reports which rows/objects were (or would be) deleted.
type PurgeOrphansOutput struct {
	// DBOrphans are video IDs whose storage_key file was not found in S3.
	DBOrphans []uuid.UUID
	// S3Orphans are S3 object keys not referenced by any video row.
	// Empty when IncludeS3Orphans was false.
	S3Orphans []string
}

// PurgeOrphans cross-references the video DB rows against S3 storage and
// removes entries that have no counterpart on the other side.
//
// Phase 1 (always): for every DB row, Stat its storage_key in S3. Any row
// whose file is missing is a DB orphan and is hard-deleted from the database.
//
// Phase 2 (opt-in): list all S3 objects and build a set of known keys from
// the DB (both storage_key and thumbnail_key). Any S3 key absent from that
// set is deleted from storage.
func (s *Service) PurgeOrphans(ctx context.Context, input PurgeOrphansInput) (*PurgeOrphansOutput, *apperror.AppError) {
	infos, err := s.videoRepo.ListAllStorageKeys(ctx)
	if err != nil {
		s.log.Error("listing video storage keys for purge", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list video storage keys", err)
	}

	out := &PurgeOrphansOutput{}

	// Phase 1 — DB rows with no matching S3 object.
	for _, info := range infos {
		obj, err := s.storage.Stat(ctx, info.StorageKey)
		if err != nil {
			s.log.Error("stat storage key during purge", "key", info.StorageKey, "error", err)
			return nil, apperror.Internal(apperror.CodeInternalError, "failed to stat storage object", err)
		}
		if obj != nil {
			continue
		}
		out.DBOrphans = append(out.DBOrphans, info.ID)
		if !input.DryRun {
			if err := s.videoRepo.Delete(ctx, info.ID); err != nil {
				s.log.Error("deleting orphaned video row", "id", info.ID, "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to delete orphaned video", err)
			}
		}
	}

	s.log.Info("purge orphans phase 1 complete",
		"db_orphans", len(out.DBOrphans),
		"dry_run", input.DryRun,
	)

	if !input.IncludeS3Orphans {
		return out, nil
	}

	// Phase 2 — S3 objects with no matching DB row.
	// Build a set of all keys the DB knows about, plus HLS directory prefixes
	// derived from each video's ID (the CDN stores HLS at videos/{id}/hls/).
	knownKeys := make(map[string]struct{}, len(infos)*2)
	knownHLSPrefixes := make([]string, 0, len(infos))
	for _, info := range infos {
		if info.StorageKey != "" {
			knownKeys[info.StorageKey] = struct{}{}
		}
		if info.ThumbnailKey != "" {
			knownKeys[info.ThumbnailKey] = struct{}{}
		}
		knownHLSPrefixes = append(knownHLSPrefixes, fmt.Sprintf("videos/%s/hls/", info.ID))
	}

	s3Keys, err := s.storage.ListObjects(ctx, "")
	if err != nil {
		s.log.Error("listing S3 objects for purge", "error", err)
		return nil, apperror.Internal(apperror.CodeInternalError, "failed to list storage objects", err)
	}

	for _, key := range s3Keys {
		if _, known := knownKeys[key]; known {
			continue
		}
		if hasKnownPrefix(key, knownHLSPrefixes) {
			continue
		}
		out.S3Orphans = append(out.S3Orphans, key)
		if !input.DryRun {
			if err := s.storage.Delete(ctx, key); err != nil {
				s.log.Error("deleting orphaned S3 object", "key", key, "error", err)
				return nil, apperror.Internal(apperror.CodeInternalError, "failed to delete orphaned storage object", err)
			}
		}
	}

	s.log.Info("purge orphans phase 2 complete",
		"s3_orphans", len(out.S3Orphans),
		"dry_run", input.DryRun,
	)

	return out, nil
}

func hasKnownPrefix(key string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
