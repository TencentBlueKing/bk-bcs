  resource.customizations.health.tkex.tencent.com_ImageLoader: |
    hs = {}
    if obj.status ~= nil then
      if obj.status.observedGeneration == nil then
        hs.status = "Progressing"
        hs.message = "Waiting for ImageLoader to finish: observed generation is empty"
        return hs
      end
      if obj.status.observedGeneration < obj.metadata.generation then
        hs.status = "Progressing"
        hs.message = "Waiting for ImageLoader to finish: observed generation less than desired generation"
        return hs
      end
      
      if obj.status.desired ~= nil then
        if obj.status.completed ~= nil then
          if obj.status.desired ~= obj.status.completed then
            hs.status = "Progressing"
            hs.message = "Waiting for ImageLoader to finish: job is running"
            return hs
          end
        end
      end

      if obj.status.desired ~= nil then
        if obj.status.succeed ~= nil then
          if obj.status.desired == obj.status.succeed then
            hs.status = "Healthy"
            return hs
          end
        end
      end
      
      if obj.status.failedStatuses ~= nil then
        if next(obj.status.failedStatuses) ~= nil then
          hs.status = "Degraded"
          hs.message = "ImageLoader failed: please check the failedStatuses"
          return hs
        end
      end
    end

    hs.status = "Progressing"
    hs.message = "Waiting for ImageLoader"
    return hs