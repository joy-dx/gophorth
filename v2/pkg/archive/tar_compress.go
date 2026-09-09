package archive

import (
	"archive/tar"
	"context"
	"fmt"
	"os"

	kpgzip "github.com/klauspost/compress/gzip"
	kpzstd "github.com/klauspost/compress/zstd"
)

func CompressTarGzip(ctx context.Context, opts *CompressOptions) error {
	if opts == nil {
		opts = DefaultCompressOptions()
	}

	out, err := os.Create(opts.Destination)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer func() {
		_ = out.Close()
		if opts.OnCancelCleanup {
			select {
			case <-ctx.Done():
				_ = os.Remove(opts.Destination)
			default:
			}
		}
	}()

	gw, err := kpgzip.NewWriterLevel(out, kpgzip.BestCompression)
	if err != nil {
		return fmt.Errorf("gzip create: %w", err)
	}
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := addFilesToTar(ctx, tw, opts); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return gw.Close()
}

func CompressTarZstd(ctx context.Context, opts *CompressOptions) error {
	if opts == nil {
		opts = DefaultCompressOptions()
	}

	out, err := os.Create(opts.Destination)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer func() {
		_ = out.Close()
		if opts.OnCancelCleanup {
			select {
			case <-ctx.Done():
				_ = os.Remove(opts.Destination)
			default:
			}
		}
	}()

	enc, err := kpzstd.NewWriter(out, kpzstd.WithEncoderLevel(kpzstd.SpeedBetterCompression))
	if err != nil {
		return fmt.Errorf("zstd writer: %w", err)
	}
	defer enc.Close()

	tw := tar.NewWriter(enc)
	defer tw.Close()

	if err := addFilesToTar(ctx, tw, opts); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return enc.Close()
}

func CompressTarRaw(ctx context.Context, opts *CompressOptions) error {
	if opts == nil {
		opts = DefaultCompressOptions()
	}

	out, err := os.Create(opts.Destination)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer func() {
		_ = out.Close()
		if opts.OnCancelCleanup {
			select {
			case <-ctx.Done():
				_ = os.Remove(opts.Destination)
			default:
			}
		}
	}()

	tw := tar.NewWriter(out)
	defer tw.Close()

	if err := addFilesToTar(ctx, tw, opts); err != nil {
		return err
	}

	return tw.Close()
}
