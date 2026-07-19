// Package images implements LXC template and ISO image management.
package images

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Adz-ai/proxmox-cli/cmd/utility"
	"github.com/Adz-ai/proxmox-cli/internal/interfaces"
	"github.com/luthermonson/go-proxmox"
	"github.com/spf13/cobra"
)

// NewTemplateCmd builds the top-level template command group.
func NewTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage LXC container templates",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newTemplateAvailableCmd(), newTemplateListCmd(), newTemplateDownloadCmd())
	return cmd
}

type applianceSummary struct {
	Template    string `json:"template"`
	OS          string `json:"os"`
	Description string `json:"description,omitempty"`
}

func newTemplateAvailableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "available",
		Short: "List templates available for download",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			node, format, err := nodeAndFormatFromFlags(cmd)
			if err != nil {
				return err
			}

			appliances, err := node.Appliances(ctx)
			if err != nil {
				return fmt.Errorf("list available templates: %w", err)
			}

			summaries := make([]applianceSummary, 0, len(appliances))
			for _, appliance := range appliances {
				description := strings.TrimSpace(appliance.Description)
				if len(description) > 60 {
					description = description[:57] + "..."
				}
				summaries = append(summaries, applianceSummary{
					Template:    appliance.Template,
					OS:          appliance.Os,
					Description: description,
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}

			fmt.Fprintf(out, "%-55s %-12s %s\n", "Template", "OS", "Description")
			fmt.Fprintf(out, "%-55s %-12s %s\n", "--------", "--", "-----------")
			for _, summary := range summaries {
				fmt.Fprintf(out, "%-55s %-12s %s\n", summary.Template, summary.OS, summary.Description)
			}
			if len(summaries) == 0 {
				fmt.Fprintln(out, "No templates available")
			}
			return nil
		},
	}

	addNodeFlag(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

type contentSummary struct {
	VolID     string `json:"volid"`
	Size      uint64 `json:"size_bytes"`
	CreatedAt string `json:"created_at,omitempty"`
}

func newTemplateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List downloaded templates on a storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			node, format, err := nodeAndFormatFromFlags(cmd)
			if err != nil {
				return err
			}
			storage, err := storageFromFlags(cmd)
			if err != nil {
				return err
			}

			templates, err := node.VzTmpls(ctx, storage)
			if err != nil {
				return fmt.Errorf("list templates on storage %q: %w", storage, err)
			}

			summaries := make([]contentSummary, 0, len(templates))
			for _, template := range templates {
				summaries = append(summaries, contentSummary{
					VolID:     template.VolID,
					Size:      uint64(template.Size),
					CreatedAt: formatCtime(uint64(template.CTime)),
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}
			printContentTable(out, fmt.Sprintf("Templates on %s:", storage), summaries)
			return nil
		},
	}

	addNodeFlag(cmd)
	addStorageFlag(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

func newTemplateDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download an appliance template to a storage",
		Long: `Start downloading a template from the appliance index, e.g.:

  proxmox-cli template download -n pve --storage local --template debian-12-standard_12.7-1_amd64.tar.zst`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			node, _, err := nodeAndFormatFromFlags(cmd)
			if err != nil {
				return err
			}
			storage, err := storageFromFlags(cmd)
			if err != nil {
				return err
			}
			template, err := cmd.Flags().GetString("template")
			if err != nil {
				return fmt.Errorf("read template flag: %w", err)
			}
			template = strings.TrimSpace(template)
			if template == "" {
				return fmt.Errorf("template cannot be empty")
			}

			upid, err := node.DownloadAppliance(ctx, template, storage)
			if err != nil {
				return fmt.Errorf("download template %q: %w", template, err)
			}

			fmt.Fprintf(out, "Download of %s started (task %s)\n", template, upid)
			fmt.Fprintln(out, "Watch progress with: proxmox-cli nodes tasks -n <node> -r")
			return nil
		},
	}

	addNodeFlag(cmd)
	addStorageFlag(cmd)
	cmd.Flags().String("template", "", "Template name from 'template available'")
	if err := cmd.MarkFlagRequired("template"); err != nil {
		panic(err)
	}
	return cmd
}

// NewISOCmd builds the top-level iso command group.
func NewISOCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iso",
		Short: "Manage ISO images",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newISOListCmd(), newISODownloadCmd())
	return cmd
}

func newISOListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ISO images on a storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			node, format, err := nodeAndFormatFromFlags(cmd)
			if err != nil {
				return err
			}
			storageName, err := storageFromFlags(cmd)
			if err != nil {
				return err
			}

			storage, err := node.Storage(ctx, storageName)
			if err != nil {
				return fmt.Errorf("get storage %q: %w", storageName, err)
			}
			content, err := storage.GetContent(ctx)
			if err != nil {
				return fmt.Errorf("list content of storage %q: %w", storageName, err)
			}

			summaries := []contentSummary{}
			for _, item := range content {
				if !strings.Contains(item.Volid, "iso/") {
					continue
				}
				summaries = append(summaries, contentSummary{
					VolID:     item.Volid,
					Size:      item.Size,
					CreatedAt: formatCtime(uint64(item.Ctime)),
				})
			}

			if format == "json" {
				return utility.PrintJSON(out, summaries)
			}
			printContentTable(out, fmt.Sprintf("ISO images on %s:", storageName), summaries)
			return nil
		},
	}

	addNodeFlag(cmd)
	addStorageFlag(cmd)
	utility.AddOutputFlag(cmd)
	return cmd
}

func newISODownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download an ISO image from a URL to a storage",
		Long: `Start downloading an ISO directly onto a Proxmox storage, e.g.:

  proxmox-cli iso download -n pve --storage local --url https://... --filename debian-12.iso`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			node, _, err := nodeAndFormatFromFlags(cmd)
			if err != nil {
				return err
			}
			storage, err := storageFromFlags(cmd)
			if err != nil {
				return err
			}
			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return fmt.Errorf("read url flag: %w", err)
			}
			filename, err := cmd.Flags().GetString("filename")
			if err != nil {
				return fmt.Errorf("read filename flag: %w", err)
			}
			url = strings.TrimSpace(url)
			filename = strings.TrimSpace(filename)
			if url == "" {
				return fmt.Errorf("url cannot be empty")
			}
			if filename == "" {
				return fmt.Errorf("filename cannot be empty")
			}

			upid, err := node.StorageDownloadURL(ctx, &proxmox.StorageDownloadURLOptions{
				Content:  "iso",
				Storage:  storage,
				URL:      url,
				Filename: filename,
			})
			if err != nil {
				return fmt.Errorf("download ISO %q: %w", filename, err)
			}

			fmt.Fprintf(out, "Download of %s started (task %s)\n", filename, upid)
			fmt.Fprintln(out, "Watch progress with: proxmox-cli nodes tasks -n <node> -r")
			return nil
		},
	}

	addNodeFlag(cmd)
	addStorageFlag(cmd)
	cmd.Flags().String("url", "", "URL to download from")
	cmd.Flags().String("filename", "", "Filename to store the ISO as")
	for _, flag := range []string{"url", "filename"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
	return cmd
}

func addNodeFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("node", "n", "", "Node name")
	if err := cmd.MarkFlagRequired("node"); err != nil {
		panic(err)
	}
	utility.RegisterNodeFlagCompletion(cmd, "node")
}

func addStorageFlag(cmd *cobra.Command) {
	cmd.Flags().String("storage", "", "Storage name")
	if err := cmd.MarkFlagRequired("storage"); err != nil {
		panic(err)
	}
}

func nodeAndFormatFromFlags(cmd *cobra.Command) (node interfaces.NodeInterface, format string, err error) {
	nodeName, err := cmd.Flags().GetString("node")
	if err != nil {
		return nil, "", fmt.Errorf("read node flag: %w", err)
	}
	nodeName = strings.TrimSpace(nodeName)
	if nodeName == "" {
		return nil, "", fmt.Errorf("node cannot be empty")
	}
	if cmd.Flags().Lookup("output") != nil {
		format, err = utility.OutputFormat(cmd)
		if err != nil {
			return nil, "", err
		}
	}

	client, err := utility.AuthenticatedClient()
	if err != nil {
		return nil, "", fmt.Errorf("authenticate Proxmox client: %w", err)
	}
	node, err = client.Node(cmd.Context(), nodeName)
	if err != nil {
		return nil, "", fmt.Errorf("get node %q: %w", nodeName, err)
	}
	return node, format, nil
}

func storageFromFlags(cmd *cobra.Command) (string, error) {
	storage, err := cmd.Flags().GetString("storage")
	if err != nil {
		return "", fmt.Errorf("read storage flag: %w", err)
	}
	storage = strings.TrimSpace(storage)
	if storage == "" {
		return "", fmt.Errorf("storage cannot be empty")
	}
	return storage, nil
}

func formatCtime(ctime uint64) string {
	if ctime == 0 {
		return ""
	}
	return time.Unix(int64(ctime), 0).UTC().Format(time.RFC3339)
}

func printContentTable(out io.Writer, header string, summaries []contentSummary) {
	fmt.Fprintln(out, header)
	fmt.Fprintln(out, strings.Repeat("=", len(header)))
	for _, summary := range summaries {
		created := summary.CreatedAt
		if created == "" {
			created = "N/A"
		}
		fmt.Fprintf(out, "%-60s %10.2f GiB  %s\n",
			summary.VolID, float64(summary.Size)/(1024*1024*1024), created)
	}
	if len(summaries) == 0 {
		fmt.Fprintln(out, "No items found")
	}
}
