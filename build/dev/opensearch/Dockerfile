FROM opensearchproject/opensearch:latest
USER root

RUN /usr/share/opensearch/bin/opensearch-plugin remove opensearch-security
RUN /usr/share/opensearch-dashboards/bin/opensearch-dashboards-plugin remove securityDashboards

EXPOSE 9200