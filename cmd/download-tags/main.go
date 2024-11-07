package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/jackc/pgx/v5"
	"k8s.io/utils/ptr"
)

const (
	OrgGroupTagKey = "OrgGroup"
	DateFormat     = "2006-01-02"
)

func main() {
	ctx := context.Background()

	db, err := pgx.Connect(ctx, "postgres://postgres:supersecret@localhost:5432/postgres")
	if err != nil {
		panic(err)
	}

	taCl, err := newTaggingAPIClient(ctx)
	if err != nil {
		panic(err)
	}

	pag := resourcegroupstaggingapi.NewGetResourcesPaginator(taCl, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key: ptr.To(OrgGroupTagKey),
			},
		},
		ResourcesPerPage: ptr.To(int32(100)),
	})

	resourceARNsToOrgGroups := make(map[string]string)
	for pag.HasMorePages() {
		log.Print("Iterating")
		data, err := pag.NextPage(ctx)
		if err != nil {
			panic(err)
		}
		for _, tagMapping := range data.ResourceTagMappingList {
			resourceARN := *tagMapping.ResourceARN
			orgGroup := extractTagValue(tagMapping.Tags, OrgGroupTagKey)
			resourceARNsToOrgGroups[resourceARN] = orgGroup
		}
	}

	log.Printf("Extracted %d resources", len(resourceARNsToOrgGroups))

	insertQuery := buildInsertQuery(time.Now(), resourceARNsToOrgGroups)
	_, err = db.Exec(ctx, insertQuery)
	if err != nil {
		panic(err)
	}

}

func newTaggingAPIClient(ctx context.Context) (*resourcegroupstaggingapi.Client, error) {
	awsCnf, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load AWS config error: %w", err)
	}
	return resourcegroupstaggingapi.NewFromConfig(awsCnf), nil
}

func extractTagValue(tags []types.Tag, tagName string) string {
	for _, tag := range tags {
		if *tag.Key == tagName {
			return *tag.Value
		}
	}
	return ""
}

func buildInsertQuery(createdAt time.Time, resourceARNsToOrgGroups map[string]string) string {
	baseQuery := "INSERT INTO resource_owners (resource_arn, org_group, created_at) VALUES \n"
	dateFormatted := createdAt.Format(DateFormat)
	values := make([]string, 0, len(resourceARNsToOrgGroups))
	for resourceARN, orgGroup := range resourceARNsToOrgGroups {
		values = append(values, fmt.Sprintf("('%s', '%s', '%s')", resourceARN, orgGroup, dateFormatted))
	}
	baseQuery += strings.Join(values, ",\n")
	baseQuery += ";"
	return baseQuery
}
