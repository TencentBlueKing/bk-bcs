/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package types defines the data structures for database models.
package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionPushTemplate is the name of the push templates collection in MongoDB.
const CollectionPushTemplate = "push_templates"

// PushTemplate represents a push notification template in the database.
type PushTemplate struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TemplateID   string             `bson:"template_id" json:"template_id"`
	Domain       string             `bson:"domain" json:"domain"`
	TemplateType string             `bson:"template_type" json:"template_type"`
	Content      TemplateContent    `bson:"content" json:"content"`
	Creator      string             `bson:"creator" json:"creator"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

// TemplateContent defines the content of a push template.
type TemplateContent struct {
	Title     string   `bson:"title" json:"title"`
	Body      string   `bson:"body" json:"body"`
	Variables []string `bson:"variables" json:"variables"`
}
