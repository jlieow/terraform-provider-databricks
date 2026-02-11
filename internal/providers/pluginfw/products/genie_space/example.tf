# Example Terraform configuration for databricks_genie_space resource
# This example demonstrates all available serialized_space options

resource "databricks_sql_endpoint" "this" {
  name             = "Genie Warehouse"
  cluster_size     = "Small"
  max_num_clusters = 1
}

# Comprehensive example with all serialized_space features
resource "databricks_genie_space" "example" {
  title       = "Revenue Analysis Space"
  description = "Ask questions about revenue data using TPC-H sample dataset"
  warehouse_id = databricks_sql_endpoint.this.id

  # Serialized space configuration (JSON string)
  # Note: IDs must be 32-character lowercase hexadecimal strings
  # Note: Arrays must be pre-sorted (tables by identifier, columns by column_name)
  serialized_space = jsonencode({
    version = 2

    # Sample questions to guide users
    config = {
      sample_questions = [
        {
          id       = "00000000000000000000000000000001"
          question = ["Show total revenue by month"]
        },
        {
          id       = "00000000000000000000000000000002"
          question = ["What are the top 10 customers by order value?"]
        },
        {
          id       = "00000000000000000000000000000003"
          question = ["Which region has the highest sales?"]
        }
      ]
    }

    # Data sources available to the space
    data_sources = {
      # Tables must be sorted alphabetically by identifier
      tables = [
        {
          identifier  = "samples.tpch.customer"
          description = ["Customer master data containing demographics and contact information"]

          # column_configs must be an ARRAY, sorted alphabetically by column_name
          column_configs = [
            {
              column_name             = "c_address"
              description             = ["Customer street address"]
              synonyms                = ["address", "street"]
              exclude                 = false
              enable_entity_matching  = false
              enable_format_assistance = false
            },
            {
              column_name             = "c_custkey"
              description             = ["Unique customer identifier", "Primary key for customer table"]
              synonyms                = ["customer_id", "customer_key", "cust_id"]
              exclude                 = false
              enable_entity_matching  = true
              enable_format_assistance = true
            },
            {
              column_name             = "c_name"
              description             = ["Customer name"]
              synonyms                = ["customer_name", "name"]
              exclude                 = false
              enable_entity_matching  = true
              enable_format_assistance = false
            }
          ]
        },
        {
          identifier  = "samples.tpch.lineitem"
          description = ["Line items for orders containing product and pricing details"]
          column_configs = [
            {
              column_name = "l_discount"
              description = ["Discount percentage applied to the line item"]
              synonyms    = ["discount", "discount_rate"]
            },
            {
              column_name = "l_extendedprice"
              description = ["Extended price before discount"]
              synonyms    = ["price", "extended_price", "line_price"]
            },
            {
              column_name = "l_quantity"
              description = ["Quantity ordered"]
              synonyms    = ["qty", "quantity", "amount"]
            }
          ]
        },
        {
          identifier  = "samples.tpch.nation"
          description = ["Nation reference data"]
        },
        {
          identifier  = "samples.tpch.orders"
          description = ["Order header information"]
          column_configs = [
            {
              column_name = "o_custkey"
              description = ["Foreign key to customer"]
              synonyms    = ["customer_key"]
            },
            {
              column_name = "o_orderdate"
              description = ["Date when order was placed"]
              synonyms    = ["order_date", "date"]
            },
            {
              column_name = "o_orderstatus"
              description = ["Order status: F=Fulfilled, O=Open, P=Pending"]
              synonyms    = ["status", "order_status"]
            },
            {
              column_name = "o_totalprice"
              description = ["Total order value"]
              synonyms    = ["total", "order_total", "total_price"]
            }
          ]
        },
        {
          identifier = "samples.tpch.part"
        },
        {
          identifier = "samples.tpch.partsupp"
        },
        {
          identifier = "samples.tpch.region"
        },
        {
          identifier = "samples.tpch.supplier"
        }
      ]

      # Optional: Metric views (same structure as tables)
      # metric_views = [
      #   {
      #     identifier  = "catalog.schema.my_metric_view"
      #     description = ["Pre-aggregated metrics"]
      #   }
      # ]
    }

    # Instructions for the LLM
    # NOTE: The instructions schema is complex and may vary by API version.
    # Start simple and add fields incrementally based on API responses.
    instructions = {
      # Text instructions - high-level guidance (max 1 per space)
      text_instructions = [
        {
          id      = "00000000000000000000000000000010"
          content = "When calculating revenue, always use l_extendedprice * (1 - l_discount) from the lineitem table. For date filtering, use o_orderdate from the orders table."
        }
      ]

      # Example questions with SQL answers
      # NOTE: question = array, sql = array
      example_question_sqls = [
        {
          id       = "00000000000000000000000000000020"
          question = ["What is the total revenue?"]
          sql      = ["SELECT SUM(l_extendedprice * (1 - l_discount)) as total_revenue FROM samples.tpch.lineitem"]
        }
      ]

      # Join specifications
      # NOTE: sql = array, comment = array
      join_specs = [
        {
          id = "00000000000000000000000000000030"
          left = {
            identifier = "samples.tpch.orders"
            alias      = "o"
          }
          right = {
            identifier = "samples.tpch.customer"
            alias      = "c"
          }
          sql     = ["o.o_custkey = c.c_custkey"]
          comment = ["Join orders to customers using customer key"]
        },
        {
          id = "00000000000000000000000000000031"
          left = {
            identifier = "samples.tpch.lineitem"
            alias      = "l"
          }
          right = {
            identifier = "samples.tpch.orders"
            alias      = "o"
          }
          sql     = ["l.l_orderkey = o.o_orderkey"]
          comment = ["Join line items to orders"]
        }
      ]

      # SQL snippets for reusable components
      sql_snippets = {
        # Predefined filters
        filters = [
          {
            id           = "00000000000000000000000000000040"
            sql          = "o.o_orderstatus = 'F'"
            display_name = "Fulfilled Orders"
            synonyms     = ["completed orders", "finished orders"]
          }
        ]

        # Predefined expressions
        expressions = [
          {
            id           = "00000000000000000000000000000050"
            alias        = "order_year"
            sql          = "YEAR(o.o_orderdate)"
            display_name = "Order Year"
            synonyms     = ["year", "fiscal year"]
          }
        ]

        # Predefined measures
        measures = [
          {
            id           = "00000000000000000000000000000060"
            alias        = "total_revenue"
            sql          = "SUM(l.l_extendedprice * (1 - l.l_discount))"
            display_name = "Total Revenue"
            synonyms     = ["revenue", "sales", "total sales"]
          }
        ]
      }
    }

    # Optional: Benchmark questions for testing
    # benchmarks = {
    #   questions = [
    #     {
    #       id       = "00000000000000000000000000000070"
    #       question = ["What is the total revenue for 1995?"]
    #       answer = [
    #         {
    #           format  = "SQL"
    #           content = ["SELECT SUM(l_extendedprice * (1 - l_discount)) FROM samples.tpch.lineitem l JOIN samples.tpch.orders o ON l.l_orderkey = o.o_orderkey WHERE YEAR(o.o_orderdate) = 1995"]
    #         }
    #       ]
    #     }
    #   ]
    # }
  })
}

# Output the space ID
output "genie_space_id" {
  value = databricks_genie_space.example.space_id
}

# Import example:
# terraform import databricks_genie_space.example <space_id>
